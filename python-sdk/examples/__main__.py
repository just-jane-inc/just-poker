# import openapi_client as poker_api
import argparse
import asyncio
import json
import os

from textual.app import App, ComposeResult
from textual.containers import Horizontal, Vertical
from textual.reactive import reactive
from textual.widgets import Footer, Static

from openapi_client.models.game_game_dto import GameGameDTO
from openapi_client.models.game_table_dto import GameTableDTO
import src.poker_helpers as help
import src.bot as bot
from openapi_client.models import GameCardDTO, GamePlayerDTO

from examples.tmp import get_test_user


parser = argparse.ArgumentParser(prog="test tui")
parser.add_argument("--game-id", type=str, required=True, help="game id")
parser.add_argument("--player-id", type=str, required=False, help="")

base_url = "http://localhost:7653"


def main():
    PokerApp(ansi_color=True).run()


class Table(Static):
    table: reactive[GameTableDTO] = reactive(GameTableDTO())
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
    players = reactive([])
    card_map = help.get_unicode_mapping()
    me: bot.PokerBot | reactive[None] = reactive(None)

    def render(self) -> str:
        view = ""
        for player in self.players:
            chip_stack = chip_sum(player.stack)
            current_bet = chip_sum(player.current_bet)
            view += f"{player.display_name}\t{chip_stack}\t({current_bet})\n"

            if player.hole:
                rank = help.CardRank(player.hole[0].rank)
                suit = help.CardSuit(player.hole[0].suit)
                card_one = self.card_map[suit][rank]

                rank = help.CardRank(player.hole[1].rank)
                suit = help.CardSuit(player.hole[1].suit)
                card_two = self.card_map[suit][rank]
                view += f"{card_one} {card_two}"

            view += "\n\n"

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
        ("c", "call", "Call"),
        ("r", "raise_bet", "Raise"),
        ("a", "all_in", "All-in"),
        ("q", "quit", "Quit"),
    ]

    def compose(self) -> ComposeResult:
        with Horizontal(id="main"):
            yield Players(id="players")
            yield Table(id="table")

        yield Footer()

    async def on_mount(self) -> None:
        game_state = None
        args = parser.parse_args()
        jane = get_test_user("jane")
        if args.game_id:
            conn = help.create_connection(base_url, jane.token)
            game_state = await help.get_game_state(conn, args.game_id)

        self.set_interval(1, self.tick)
        player = self.query_one(Players)
        player.players = game_state.table.players
        table = self.query_one(Table)
        table.table = game_state.table
        table.seconds_remaining = 100
        player.me = bot.PokerBot(base_url, jane.token, jane.user_id, args.game_id)

    def tick(self) -> None:
        table = self.query_one(Table)

        if table.seconds_remaining > 0:
            table.seconds_remaining -= 1

    async def action_fold(self) -> None:
        players = self.query_one(Players)
        await players.me.send_action("fold", {})
        self.notify("folded!")

    async def action_call(self) -> None:
        self.notify("call selected")

    def action_raise_bet(self) -> None:
        self.notify("raise selected")

    async def action_all_in(self) -> None:
        players = self.query_one(Players)
        await players.me.send_action("all_in", {})
        self.notify("all in!")


if __name__ == "__main__":
    main()
