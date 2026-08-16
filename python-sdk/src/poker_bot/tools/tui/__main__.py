# import openapi_client as poker_api
import argparse
import asyncio
import logging
import os
from collections.abc import Callable

from dotenv import load_dotenv
from textual import on
from textual.app import App, ComposeResult
from textual.containers import Container, Horizontal
from textual.reactive import reactive
from textual.screen import ModalScreen
from textual.widgets import Button, Footer, Input, Label, Static

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.models import GameGameDTO, GamePlayerDTO, GameTableDTO
from poker_bot.bot.bot import PokerBot
from poker_bot.bot.event_hub import Event, EventHub, EventType
from poker_bot.tools.tui.setup_tui_example import get_test_user, setup

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser(
    prog="just__poker: TUI", description="a TUI that can be used to play a game of poker on the server"
)

parser.add_argument("--game-id", "--id", type=str, required=False, help="game id")
parser.add_argument("--user", type=str, required=False, help="")
parser.add_argument("--setup", type=int, default=0, required=False, help="")

logger = logging.getLogger("tui")

if os.name == "nt":
    # For windows only tool usage, fix for cert stuff
    try:
        import truststore

        truststore.inject_into_ssl()
    finally:
        pass


def main():
    args = parser.parse_args()
    if args.setup:
        asyncio.run(setup(base_url, args.setup))
    else:
        PokerApp(ansi_color=True).run()


class InputPopup(ModalScreen[str]):
    """A modal screen with a text input box."""

    DEFAULT_CSS = """
    InputPopup {
        align: center middle;
    }
    InputPopup > Container {
        width: 40;
        height: auto;
        background: $surface;
        border: thick $background 80%;
        padding: 1 2;
    }
    InputPopup Label {
        margin-bottom: 1;
    }
    InputPopup Input {
        width: 30%;
    }
    """

    def compose(self) -> ComposeResult:
        yield Label("raise to: ")
        yield Input(placeholder="...", id="text-box")
        yield Button("Submit", variant="primary", id="ok")

    @on(Button.Pressed, "#ok")
    async def submit_text(self) -> None:
        text_input = self.query_one("#text-box", Input)
        self.dismiss(text_input.value)

    @on(Input.Submitted)
    async def enter_submitted(self, event: Input.Submitted) -> None:
        self.dismiss(event.value)


class Table(Static):
    table: reactive[GameTableDTO | None] = reactive(GameTableDTO(), layout=True)
    card_map = help.get_unicode_mapping()
    winner: str | None = None
    game_id: str | None = None

    def render(self) -> str:
        # -ˋˏ ༻❁❀༺ ˎˊ-
        # ˖⁺‧₊˚˚₊‧⁺˖
        # .𓋼𓍊 𓆏 𓍊𓋼𓍊.☆
        view = f"♠♥ ☆༻❁♡✿⊱༻{self.game_id!s:^6}༺⊰✿♡❁༺☆ ♦♣\n"
        view += f"║|{'-' * 22}|║\n"
        if self.winner:
            view += f"║| ♠♥{self.winner[:16]:^16}♦♣ |║\n"
        if self.table is None or self.table.street is None:
            return view

        to_call = self.table.current_round.bet

        if not self.winner:
            view += f"║| POT: {chip_sum(self.table.pot):>15} |║\n"
            view += f"║| STREET: {self.table.current_round.current_round_type.upper():>12} |║\n"
            view += f"║| TO CALL: {to_call:>11} |║\n"
            view += f"║|{'-' * 22}|║\n"
        cards_view = ""
        for card in self.table.street:
            rank = help.CardRank(card.rank)
            suit = help.CardSuit(card.suit)
            cards_view += self.card_map[suit][rank] + " "
        view += f"║| {cards_view:^20} |║\n"
        view += f"║|{'-' * 22}|║\n"

        return view


class Players(Static):
    players: reactive[list[GamePlayerDTO] | None] = reactive(None, layout=True)
    current_turn: reactive[int] = reactive(-1, layout=True)
    card_map = help.get_unicode_mapping()
    me: PokerBot | reactive[None] = reactive(None, layout=True)

    def render(self) -> str:
        view = ""
        if not self.players:
            return ""

        for player in self.players:
            chip_stack = chip_sum(player.stack)
            current_bet = chip_sum(player.current_bet)

            name = player.display_name[:8]
            if player.user_id == self.me._user_id:
                name = f"<<{name}>>"

            card_view = ""
            if player.hole:
                rank = help.CardRank(player.hole[0].rank)
                suit = help.CardSuit(player.hole[0].suit)
                card_one = self.card_map[suit][rank]

                rank = help.CardRank(player.hole[1].rank)
                suit = help.CardSuit(player.hole[1].suit)
                card_two = self.card_map[suit][rank]
                card_view += f"{card_one} {card_two}"

            view += f"{name:<10}\n{player.state:<8} {card_view:^4} {current_bet:>6}◉ {chip_stack:<6}⛀⛁\n"

            view += "\n"

        return view


def chip_sum(stack: dict[str, int]) -> int:
    return sum((int(k) * v for k, v in stack.items()))


class PokerApp(App):
    CSS = """
    Screen {
        layout: vertical;
    }

    #main {
        height: 1fr;
    }

    #players {
        padding-right: 3;
    }

    #playersbox {
        width: 1fr;
        height: 100%;
        overflow-y: auto;
    }

    #table {
        height: 100%;
        width: 1.5fr;
        padding-top: 2;
        padding-left: 5;
    }
    """

    BINDINGS = [
        ("f", "fold", "Fold"),
        ("-", "check", "Check"),
        ("c", "call", "Call"),
        ("t", "ante", "Ante"),
        ("r", "raise_bet", "Raise"),
        ("a", "all_in", "All-in"),
        ("q", "quit", "Quit"),
    ]

    game_state: GameGameDTO | None = None
    events: EventHub | None = None
    _unsubscribe: Callable[[], None] | None = None

    def compose(self) -> ComposeResult:
        with Horizontal(id="main"):
            with Container(id="playersbox"):
                yield Players(id="players")
            yield Table(id="table")

        yield Footer()

    async def on_mount(self) -> None:
        args = parser.parse_args()
        self.game_state = None
        user = get_test_user(args.user)

        me = PokerBot(base_url, user.token, user.user_id, args.game_id, timeout=5)
        self.query_one(Players).me = me

        self.events = me.events

        if self.events:
            # Example of using as a decorator with specific event types
            @self.events.on_event(
                EventType.WELCOME,
                EventType.GAME_STATE_UPDATE,
                # EventType.STARTING_GAME,
                EventType.GAME_OVER,
            )
            async def _on_update(event: Event) -> None:
                logger.debug(f"received for Player [{me._user_id}]:\t{event.event_type} - {event.data}")
                if isinstance(event.data, GameGameDTO):
                    await self.apply_state(event.data)
                elif event.event_type == EventType.GAME_OVER:
                    await self.apply_game_over(event.data)

            # alternative - subscribe inline with reference hook
            # sub = self.events.subscribe(EventType.GAME_OVER, self.apply_game_over)

            self._unsubscribe = _on_update.unsubscribe
            # start it anyways
            await self.events.start()

        # Fetch initial via POST if desired, but welcome msg on websocket should return it.
        # state = await me.get_game_state()
        # if state is not None and self.game_state is None:
        #     await self.apply_state(state)

    async def apply_game_over(self, data: list[GamePlayerDTO]) -> None:
        if data is None:
            return
        self.query_one(Players).players = data
        self.query_one(Table).table.pot.clear()

        for player in data:
            if player.stack is not None and chip_sum(player.stack) > 0:
                self.query_one(Table).winner = player.display_name
                break

    async def apply_state(self, state: GameGameDTO) -> None:
        self.game_state = state
        if state.table is not None:
            t = self.query_one(Table)
            t.table = state.table
            t.game_id = state.id
            self.query_one(Players).players = state.table.players

    async def on_unmount(self) -> None:
        if self._unsubscribe is not None:
            self._unsubscribe()
        if self.events is not None:
            await self.events.stop()

    async def action_fold(self) -> None:
        players = self.query_one(Players)
        try:
            if not await players.me.fold():
                self.notify("fold action has timed out")
            else:
                self.notify("folded!")
        except api.ApiException as e:
            poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
            self.notify(f"encountered error [{poker_error.data.error_code}] {poker_error.data.error}")

    async def action_call(self) -> None:
        players = self.query_one(Players)
        try:
            if not await players.me.call():
                self.notify("call action has timed out")
            else:
                self.notify("call!")
        except api.ApiException as e:
            poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
            self.notify(f"encountered error [{poker_error.data.error_code}] {poker_error.data.error}")

    async def action_raise_bet(self) -> None:
        async def check_result(result: str | None) -> None:
            if result is not None:
                players = self.query_one(Players)
                try:
                    if not await players.me.raise_bet(int(result)):
                        self.notify("raise action has timed out")
                    else:
                        self.notify("raied!")
                except api.ApiException as e:
                    poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
                    self.notify(f"encountered error [{poker_error.data.error_code}] {poker_error.data.error}")

            else:
                self.query_one("#info", Static).update("Cancelled")

        self.push_screen(InputPopup(), check_result)

    async def action_ante(self) -> None:
        players = self.query_one(Players)
        try:
            if not await players.me.ante():
                self.notify("ante action has timed out")
            else:
                self.notify("ante!")
        except api.ApiException as e:
            poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
            self.notify(f"encountered error [{poker_error.data.error_code}] {poker_error.data.error}")

    async def action_check(self) -> None:
        players = self.query_one(Players)
        try:
            if not await players.me.check():
                self.notify("check action has timed out")
            else:
                self.notify("check!")
        except api.ApiException as e:
            poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
            self.notify(f"encountered error [{poker_error.data.error_code}] {poker_error.data.error}")

    async def action_all_in(self) -> None:
        players = self.query_one(Players)
        try:
            if not await players.me.all_in():
                self.notify("all-in action has timed out")
            else:
                self.notify("all-in!")
        except api.ApiException as e:
            poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
            self.notify(f"encountered error [{poker_error.data.error_code}] {poker_error.data.error}")


if __name__ == "__main__":
    main()
