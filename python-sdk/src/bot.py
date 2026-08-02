import asyncio
from dataclasses import dataclass
import logging
from os import wait

import httpx

import src.poker_exceptions as ex
import src.poker_helpers as help
from openapi_client import *

logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger("bot")


@dataclass
class Chips:
    denomination: int
    count: int


def sum_chips(chip_stack: list[Chips]) -> int:
    sum = 0
    for chips in chip_stack:
        sum += chips.denomination * chips.count

    return sum


def convert_stack(stack) -> list[Chips]:
    chips: list[Chips] = []
    for d, c in stack.items():
        chips.append(Chips(int(d), c))

    return chips


def convert_chips(chips: list[Chips]) -> dict[str, int]:
    stack: dict[str, int] = {}
    for chip in chips:
        stack[str(chip.denomination)] = chip.count

    return stack


class PokerBot:
    def __init__(self, base_url: str, token: str, user_id: str, game_id: str):
        """A poker bot

        Args:
            base_url: the url for the poker server to connect to
            token: the user authorization token to use for this bot
            game_id: the id of the game to interact with

        Raises:
            CustomException: if an argument is not provided
        """
        if not token:
            raise ex.CustomException("api_key not provided")

        if not base_url:
            raise ex.CustomException("base url not provided")

        if not game_id:
            raise ex.CustomException("game id not provided")

        if not user_id:
            raise ex.CustomException("user id not provided")

        self._game_id = game_id
        self._api_client = help.create_connection(base_url, token)
        self._user_api = UserApi(self._api_client)
        self._game_api = GameApi(self._api_client)
        self._joined = False
        self._player = None
        self._user_id = user_id
        self._current_stack: list[Chips] = []

    async def join_game(self):
        if self._joined:
            return

        _ = await self._game_api.game_game_id_player_post(self._game_id)

    async def start_game(self):
        await self._game_api.game_game_id_started_post(self._game_id)

    async def get_game_state(self) -> GameGameDTO | None:
        resp = await self._game_api.game_game_id_state_get(self._game_id)

        if resp.data is None:
            return None

        # TODO: get the LSP to stop being insane
        if resp.data.table is None or resp.data.table.players is None:
            raise ex.CustomException("what")

        for player in resp.data.table.players:
            if player.user_id == self._user_id:
                self._player = player
                self._current_stack = sorted(
                    convert_stack(self._player.stack), key=lambda i: i.denomination, reverse=True
                )
                break

        return resp.data

    async def exchange_chips(self, give: list[Chips], receive: list[Chips]):
        give_stack = {str(s.denomination): s.count for s in give}
        receive_stack = {str(s.denomination): s.count for s in receive}
        dto = GameChipExchangeDTO(give=give_stack, receive=receive_stack)
        req = GameGameIdChipExchangePostRequest(dto)
        resp = await self._game_api.game_game_id_chip_exchange_post(self._game_id, req)
        if resp.type == "error":
            raise ex.CustomException("error in chip exchange: %s", resp.data.error)

    async def break_chip(self, value: int, denominations: list[int]) -> list[Chips]:
        result: list[Chips] = []
        remaining = value

        for denom in reversed(denominations):
            if denom >= value:
                continue

            count = remaining // denom
            remaining = remaining % denom

            if count:
                result.append(Chips(denom, count))

            if remaining == 0:
                break

            if remaining:
                raise ValueError(f"cant break {value} :(")

        return result

    def merge_stack(self, chips: Chips):
        for s in self._current_stack:
            if s.denomination == chips.denomination:
                s.count += chips.count
                return

        self._current_stack.append(chips)

    async def try_cover_bet(self, amount_needed: int, bet: list[Chips]):
        denominations = [500, 100, 50, 25, 10]
        if sum(s.denomination * s.count for s in self._current_stack) <= amount_needed:
            return self._current_stack

        denominations = sorted(denominations)
        for s in self._current_stack:
            take = min(s.count, amount_needed // s.denomination)
            if take:
                logger.debug(f"taking {take}x{s.denomination} from stack for bet")
                s.count -= take
                bet.append(Chips(s.denomination, take))
                amount_needed -= take * s.denomination
            if amount_needed == 0:
                return

        for s in sorted(self._current_stack, key=lambda q: q.denomination, reverse=True):
            if s.denomination > 0 and s.denomination > amount_needed:
                logger.debug(f"exchanging 1x{s.denomination} for smaller chips")
                s.count -= 1
                broken = await self.break_chip(s.denomination, denominations)

                for b in broken:
                    self.merge_stack(b)

                return await self.try_cover_bet(amount_needed, bet)

        raise ValueError("FRICK YOU poor bozo, cant make the change")

    async def raise_bet(self, raise_to: int):
        await self.wait_for_turn()

        current_bet = convert_stack(self._player.current_bet)
        current_bet = sum(s.denomination * s.count for s in current_bet)
        raise_to = raise_to - current_bet

        bet: list[Chips] = []
        await self.try_cover_bet(raise_to, bet)
        stack = convert_chips(bet)

        await self.send_action("raise", stack)

    async def ante(self):
        state = await self.wait_for_turn()

        amount = state.table.current_round.bet
        if not amount:
            raise ex.CustomException("erm")

        bet: list[Chips] = []
        await self.try_cover_bet(amount, bet)
        stack = convert_chips(bet)

        await self.send_action("ante", stack)

    async def call(self):
        state = await self.wait_for_turn()

        amount = state.table.current_round.bet
        if not amount:
            raise ex.CustomException("erm")

        stack = convert_stack(self._player.current_bet)
        current_bet = sum(s.denomination * s.count for s in stack)
        amount -= current_bet
        bet: list[Chips] = []
        await self.try_cover_bet(amount, bet)
        stack = convert_chips(bet)

        await self.send_action("call", stack)

    async def send_action(self, intent: str, bet: dict[str, int] | None):
        dto = GamePlayerActionDTO(chips=bet, intent=intent)
        req = GameGameIdActionPostRequest(dto)
        await self._game_api.game_game_id_action_post(self._game_id, req)

    async def wait_for_turn(self, wait: int = 1) -> GameGameDTO | None:
        state = await self.get_game_state()
        if state.table.current_round.current_player_position == self._player.position:
            return state

        await asyncio.sleep(1)
