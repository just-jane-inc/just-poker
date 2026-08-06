# import openapi_client as poker_api
import argparse
import asyncio
import os

from dotenv import load_dotenv
from textual import on
from textual.app import App, ComposeResult
from textual.containers import Horizontal
from textual.reactive import reactive
from textual.screen import ModalScreen
from textual.widgets import Button, Footer, Input, Label, Static

import poker_bot.bot.poker_helpers as help
from openapi_client.models import GameGameDTO, GamePlayerDTO, GameTableDTO
from poker_bot.bot.bot import PokerBot
from poker_bot.bot.websocket_events import (
    WebSocketListener,
)
from poker_bot.tools.tui.setup_tui_example import get_test_user, setup

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser(prog="test tui")
parser.add_argument("--game-id", type=str, required=False, help="game id")
parser.add_argument("--player-name", type=str, required=False, help="")
parser.add_argument("--setup", action="store_true", required=False, help="")

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
        asyncio.run(setup(base_url))
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
    table: reactive[GameTableDTO | None] = reactive(GameTableDTO())
    card_map = help.get_unicode_mapping()
    seconds_remaining = reactive(0)

    def render(self) -> str:
        view = "====TABLE====\n"

        if self.table is None or self.table.street is None:
            return view

        to_call = self.table.current_round.bet
        view += f"POT: {chip_sum(self.table.pot)} STREET: {self.table.current_round.current_round_type.upper()} TO CALL: {to_call}"
        view += f"  [{self.seconds_remaining}]"
        view += "\n\n"
        for card in self.table.street:
            rank = help.CardRank(card.rank)
            suit = help.CardSuit(card.suit)
            view += self.card_map[suit][rank] + " "

        return view


class Players(Static):
    players: reactive[list[GamePlayerDTO] | None] = reactive(None)
    current_turn: reactive[int] = reactive(-1)
    card_map = help.get_unicode_mapping()
    me: PokerBot | reactive[None] = reactive(None)

    def render(self) -> str:
        view = ""
        if not self.players:
            return ""

        for player in self.players:
            chip_stack = chip_sum(player.stack)
            current_bet = chip_sum(player.current_bet)

            name = player.display_name
            if player.user_id == self.me._user_id:
                name = f"{name}<"

            if player.state == "active":
                view += f"* {name}\t{chip_stack}\t({current_bet})\n"
            else:
                view += f"{player.state} {name}\t{chip_stack}\t({current_bet})\n"

            if player.hole:
                rank = help.CardRank(player.hole[0].rank)
                suit = help.CardSuit(player.hole[0].suit)
                card_one = self.card_map[suit][rank]

                rank = help.CardRank(player.hole[1].rank)
                suit = help.CardSuit(player.hole[1].suit)
                card_two = self.card_map[suit][rank]
                view += f"{card_one} {card_two}"

            view += "\n\n"

        view += f"\nlength: {len(self.players)}"
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
        width: 32;
        min-width: 26;
        padding: 1 2;
        border-right: solid $primary;
    }

    #table {
        width: 1fr;
        padding: 1 3;
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
    listener: WebSocketListener | None = None

    def compose(self) -> ComposeResult:
        with Horizontal(id="main"):
            yield Players(id="players")
            yield Table(id="table")

        yield Footer()

    async def action_raise_bet(self) -> None:
        async def check_result(result: str | None) -> None:
            if result is not None:
                players = self.query_one(Players)
                await players.me.raise_bet(int(result))
            else:
                self.query_one("#info", Static).update("Cancelled")

        self.push_screen(InputPopup(), check_result)

    async def on_mount(self) -> None:
        args = parser.parse_args()
        self.game_state = None
        user = get_test_user(args.player_name)

        me = PokerBot(base_url, user.token, user.user_id, args.game_id)
        self.query_one(Players).me = me

        self.listener = me.websocket_listener()
        if self.listener is not None:
            # Example of using as a decorator with specific event types
            # @self.listener.on_event(WebSocketEventType.WELCOME, WebSocketEventType.GAME_STATE_UPDATE)
            # async def _on_update(event: WebSocketEvent) -> None:
            #     await self.apply_state(event.data)

            # alternative - using the specific game state listener
            self.listener.on_game_state(self.apply_state)
            await self.listener.start()

        # Fetch initial via POST if desired, but welcome msg on websocket should return it.
        # state = await me.get_game_state()
        # if state is not None and self.game_state is None:
        #     await self.apply_state(state)

    async def apply_state(self, state: GameGameDTO) -> None:
        self.game_state = state
        self.query_one(Table).table = state.table
        self.query_one(Players).players = state.table.players

    async def on_unmount(self) -> None:
        if self.listener is not None:
            await self.listener.stop()

    async def action_fold(self) -> None:
        players = self.query_one(Players)
        await players.me.send_action("fold", {})
        self.notify("folded!")

    async def action_call(self) -> None:
        players = self.query_one(Players)
        await players.me.call()
        self.notify("call!")

    async def action_ante(self) -> None:
        players = self.query_one(Players)
        await players.me.ante()
        self.notify("ante!")

    async def action_check(self) -> None:
        players = self.query_one(Players)
        await players.me.send_action("check", {})
        self.notify("check!")

    async def action_all_in(self) -> None:
        players = self.query_one(Players)
        await players.me.send_action("all_in", {})
        self.notify("all in!")


if __name__ == "__main__":
    main()
