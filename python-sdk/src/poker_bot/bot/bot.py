import asyncio
import logging
from dataclasses import dataclass

import poker_bot.bot.poker_exceptions as ex
import poker_bot.bot.poker_helpers as help
import poker_bot.bot.websocket_events as ws
from openapi_client import (
    GameApi,
    GameChipExchangeDTO,
    GameGameDTO,
    GameGameIdActionPostRequest,
    GameGameIdChipExchangePostRequest,
    GamePlayerActionDTO,
    UserApi,
)
from openapi_client.models.game_player_dto import GamePlayerDTO
from openapi_client.models.game_player_intent import GamePlayerIntent
from poker_bot.bot.websocket_events import (
    WebSocketEvent,
    WebSocketListener,
    WebSocketStream,
)

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
        self._base_url = base_url
        self._token = token
        self._api_client = help.create_connection(base_url, token)
        self._user_api = UserApi(self._api_client)
        self._game_api = GameApi(self._api_client)
        self._joined = False
        self._user_id = user_id
        self._player: GamePlayerDTO | None = None
        self._current_stack: list[Chips] = []
        self._current_state: GameGameDTO | None = None
        self._listener: ws.WebSocketListener | None = None

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

        self._ingest_state(resp.data)
        return resp.data

    async def connect_game_state_listener(self):
        if self._listener is not None:
            raise ex.CustomException("already have a listener")

        self._listener = self.websocket_listener()

        if self._listener is None:
            return

        @self._listener.on_event(
            ws.WebSocketEventType.WELCOME, ws.WebSocketEventType.GAME_STATE_UPDATE
        )
        async def thing(e: ws.WebSocketEvent) -> None:
            self._ingest_state(e.data)

        await self._listener.start()

    def _ingest_state(self, state: GameGameDTO):
        # TODO: get the LSP to stop being insane
        if state is None or state.table is None or state.table.players is None:
            logger.warning("state is none in _ingest_state")
            return

        for player in state.table.players:
            if player.user_id == self._user_id:
                self._player = player
                self._current_stack = sorted(
                    convert_stack(self._player.stack),
                    key=lambda i: i.denomination,
                    reverse=True,
                )
                break

        self._current_state = state

    def _ingest_update(self, event: WebSocketEvent):
        # Note: not doing anything else from data stream atm, mainly placeholder
        if event.data is not None and isinstance(event.data, GameGameDTO):
            self._ingest_state(event.data)

    def websocket_stream(self, **kwargs) -> WebSocketStream:
        return WebSocketStream(
            self._base_url,
            self._token,
            self._game_id,
            on_state=self._ingest_update,
            **kwargs,
        )

    def websocket_listener(self, **kwargs) -> WebSocketListener | None:
        if self._listener is not None:
            return self._listener
        self._listener = WebSocketListener(
            self._base_url,
            self._token,
            self._game_id,
            on_game_state=self._ingest_state,
            **kwargs,
        )
        return self._listener

    async def exchange_chips(self, give: list[Chips], receive: list[Chips]):
        give_stack = {str(s.denomination): s.count for s in give}
        receive_stack = {str(s.denomination): s.count for s in receive}
        dto = GameChipExchangeDTO(give=give_stack, receive=receive_stack)
        req = GameGameIdChipExchangePostRequest(dto)
        resp = await self._game_api.game_game_id_chip_exchange_post(self._game_id, req)
        if resp.type == "error":
            raise ex.CustomException("error in chip exchange: %s", resp.data.error)

    async def break_chip(self, value: int, denominations: list[int]) -> list[Chips]:
        """breaks a single chip with denomination of "value" into a collection of smaller chips for an exchange"""
        result: list[Chips] = []

        # the running value of chips left to breakdown
        remaining = value

        # we want to break down chips from the largest denomination down
        # reversed returns a sorted list in decending order
        for denom in reversed(denominations):
            if denom >= value:
                continue

            count = remaining // denom
            remaining = remaining % denom

            if count > 0:
                result.append(Chips(denom, count))

            if remaining == 0:
                break

        if remaining != 0:
            logger.error(
                f"{value} can not be subdivided with provided denominations {denominations}"
            )

            raise ex.CustomException(
                "invalid chip exchange request with current denominations"
            )

        return result

    def merge_stack(self, chips: Chips):
        """joins provided chip with the bots current stack"""
        for s in self._current_stack:
            if s.denomination == chips.denomination:
                s.count += chips.count
                return

        self._current_stack.append(chips)

    async def try_cover_bet(self, amount_needed: int, bet: list[Chips]):
        """attempts to construct a set of chips that satisfies the needed amount from available chip stack

        - will preform exchanges of chips as it tries to compute an amount needed
        - will return the entire stack if the amount needed exceeds available chips (all in)
        """
        logger.debug(
            f"try cover bet for [{amount_needed}] current bet: [{bet}] stack: {self._current_stack}"
        )

        # TODO: get rid of this hardcoding of denominations, the rack
        # should be determined by the game and available via a query
        denominations = [500, 100, 50, 10]
        if sum(s.denomination * s.count for s in self._current_stack) <= amount_needed:
            bet = self._current_stack
            return  # all in

        denominations = sorted(denominations)
        for s in self._current_stack:
            if s.count == 0:
                continue

            # we take the a number of chips from this denomination
            # that does not exceed our count and reduces the amount
            # of chips needed for the bet
            take = min(s.count, amount_needed // s.denomination)
            if take == 0:
                continue

            logger.debug(f"taking {take}x{s.denomination} from stack for bet")
            s.count -= take
            bet.append(Chips(s.denomination, take))
            amount_needed -= take * s.denomination

            if amount_needed == 0:
                return

        if amount_needed == 0:
            return

        # if we fail to make a bet that satisfies the provided amount_needed
        # the code below attempts to break down our remaining chips to make
        # the thing work
        sort_method = lambda q: q.denomination
        for s in sorted(self._current_stack, key=sort_method, reverse=True):
            if s.count <= 0:
                continue

            if s.denomination <= amount_needed:
                break

            logger.debug(f"exchanging 1x{s.denomination} for smaller chips")
            broken = await self.break_chip(s.denomination, denominations)
            await self.exchange_chips(give=[Chips(s.denomination, 1)], receive=broken)
            s.count -= 1

            for b in broken:
                self.merge_stack(b)

            return await self.try_cover_bet(amount_needed, bet)

        raise ValueError("could not construct a valid bet for provided amount")

    async def check(self):
        await self.wait_for_my_turn()
        await self.send_action(GamePlayerIntent.PlayerIntentCheck, {})

    async def all_in(self):
        await self.wait_for_my_turn()
        await self.send_action(GamePlayerIntent.PlayerIntentAllIn, {})

    async def raise_bet(self, raise_to: int):
        await self.wait_for_my_turn()
        current_bet = convert_stack(self._player.current_bet)
        current_bet = sum(s.denomination * s.count for s in current_bet)
        raise_to = raise_to - current_bet

        bet: list[Chips] = []
        await self.try_cover_bet(raise_to, bet)
        stack = convert_chips(bet)

        await self.send_action(GamePlayerIntent.PlayerIntentRaise, stack)

    async def ante(self):
        await self.wait_for_my_turn()
        amount = self._current_state.table.current_round.bet
        if not amount:
            raise ex.CustomException("erm")

        bet: list[Chips] = []
        await self.try_cover_bet(amount, bet)
        stack = convert_chips(bet)

        await self.send_action(GamePlayerIntent.PlayerIntentAnte, stack)

    async def call(self):
        await self.wait_for_my_turn()
        amount = self._current_state.table.current_round.bet
        if not amount:
            raise ex.CustomException("erm")

        stack = convert_stack(self._player.current_bet)
        current_bet = sum(s.denomination * s.count for s in stack)
        amount -= current_bet
        bet: list[Chips] = []
        await self.try_cover_bet(amount, bet)
        stack = convert_chips(bet)

        await self.send_action(GamePlayerIntent.PlayerIntentCall, stack)

    async def fold(self):
        await self.wait_for_my_turn()
        await self.send_action(GamePlayerIntent.PlayerIntentFold)

    async def send_action(self, intent: str, bet: dict[str, int] | None = None):
        """send an action to the joined game

        Gee, I sure can't wait for bots to hop in my pool! - ty999999

        raises:
            openapi_client.exceptions.ApiException - on any failure. exception body
            may be parseable to an openapi_client.JustResponseMessageJustErrorDTO containing
            recovery information. please see documentation on error recovery, that exists and
            is impressively verbose. surely...
        """

        if bet is None:
            bet = {}

        dto = GamePlayerActionDTO(chips=bet, intent=intent)
        req = GameGameIdActionPostRequest(dto)
        await self._game_api.game_game_id_action_post(self._game_id, req)

    async def wait_for_my_turn(self):
        if self._listener is None:
            await self.connect_game_state_listener()

        while not self.is_my_turn():
            await asyncio.sleep(0.5)

    def is_my_turn(self) -> bool:
        if self._current_state is None:
            return False

        if self._current_state.table is None:
            return None

        current_player_position = (
            self._current_state.table.current_round.current_player_position
        )
        return current_player_position == self._player.position
