import asyncio
import logging
from dataclasses import dataclass

import openapi_client as api
import poker_bot.poker_exceptions as ex
import poker_bot.poker_helpers as help
import poker_bot.websocket_events as ws
from openapi_client import GamePlayerIntent, JustResponseMessageAny
from poker_bot.event_hub import (
    EventHub,
    EventSubscriber,
    EventType,
    websocket_event_hub,
)
from poker_bot.websocket_events import (
    WebSocketEvent,
)

logger = logging.getLogger("bot")

MAX_CHIP_DOWNS = 16


# TODO: move this into helpers or whatever
@dataclass
class BetResult:
    Bet: dict[int, int]
    Give: dict[int, int] | None
    Receive: dict[int, int] | None


class PokerBot:
    def __init__(self, base_url: str, token: str, user_id: str, game_id: str, timeout: float = -1.0):
        """A poker bot

        Args:
            base_url: the url for the poker server to connect to
            token: the user authorization token to use for this bot
            game_id: the id of the game to interact with
            timeout: the number of seconds that the bot should wait in the wait_for_my_turn function

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
        self._user_api = api.UserApi(self._api_client)
        self._game_api = api.api.GameApi(self._api_client)
        self._joined = False
        self._user_id = user_id
        self._player: api.GamePlayerDTO | None = None
        self._current_stack: list[help.Chips] = []
        self._current_state: api.GameGameDTO | None = None
        self._listener: ws.WebSocketListener | None = None
        self._hub: EventHub | None = None
        self._state_subscription: EventSubscriber | None = None
        self._timeout = timeout

    def chip_total(self) -> int:
        total = 0
        for denom, count in self._player.stack.items():
            total += int(denom) * count

        return total

    def get_game_api(self) -> api.GameApi:
        return api.GameApi(self._api_client)

    async def join_game(self):
        """joins the configured game for this bot - if the bot is already joined to the game will noop"""
        if self._joined:
            return

        _ = await self._game_api.game_game_id_player_post(self._game_id)
        self._joined = True

    async def start_game(self):
        """starts the game configured for this bot"""
        await self._game_api.game_game_id_started_post(self._game_id)

    async def get_game_state(self) -> api.GameGameDTO | None:
        """gets the current game state

        when getting the game state this will also update the internally tracked game state.
        """
        resp = await self._game_api.game_game_id_state_get(self._game_id)

        if resp.data is None:
            return None

        self._ingest_game_dto(resp.data)
        return resp.data

    @property
    def events(self) -> EventHub:
        """gets reference to EventHub configured for the bot

        this will create a hub if one does not already exist.
        """
        if self._hub:
            return self._hub

        # TODO: why do we allow this to be None? it is bad to just make one in constructor?
        self._hub = websocket_event_hub(self._base_url, self._token, self._game_id)

        if self._hub is None:
            raise ex.CustomException("could not create eventhub")

        # Listen to gamestate updates immediately to bake in helpers
        # Requires something to call x.events somehow for setup, many paths automatically do it but not all
        self._state_subscription = self._hub.subscribe(
            (
                EventType.WELCOME,
                EventType.GAME_STATE_UPDATE,
                EventType.STARTING_GAME,
            ),
            self._ingest_websocket_event,
        )

        return self._hub

    async def start_events(self) -> EventHub:
        """Only necessary to call if manually setting up all listeners"""
        return await self.events.start()

    async def stop_events(self) -> None:
        """stops the EventHub

        Raises:
            ex.CustomException: if their is no hub initialized
        """
        if self._hub is None:
            raise ex.CustomException("you have made an error")

        await self._hub.stop()

    # TODO: it might be nice to link to swagger stuff for these
    async def exchange_chips(self, give: list[help.Chips], receive: list[help.Chips]):
        """exchanges chips held by the bot with the games exchange

        Args:
            give: the chips the bot is giving to the server exchange
            receive: the chips the bot wants to receive as a result of the exchange

        Raises:
            ex.CustomException: raises error if the exchange is determined to be invalid by the server
        """
        give_stack = {str(s.denomination): s.count for s in give}
        receive_stack = {str(s.denomination): s.count for s in receive}
        dto = api.GameChipExchangeDTO(give=give_stack, receive=receive_stack)
        resp = await self._game_api.game_game_id_chip_exchange_post(self._game_id, dto)
        if resp.type == "error":
            raise ex.CustomException("error in chip exchange: %s", resp.data.error)

    def merge_stack(self, chips: help.Chips):
        """joins provided chip with the bots current stack"""
        for s in self._current_stack:
            if s.denomination == chips.denomination:
                s.count += chips.count
                return

        self._current_stack.append(chips)

    async def check(self) -> bool:
        """sends the check action after waiting for the bots turn

        Returns:
            a flag indicating True if the action was actually send, a False indicating
            that the bot reached its configured timeout prior to preforming an action
        """
        await self.wait_for_my_turn()
        if not self.is_my_turn():
            return False

        return await self.send_action(api.GamePlayerIntent.PlayerIntentCheck, {})

    async def all_in(self):
        """sends the all in action after waiting for the bots turn

        Returns:
            a flag indicating True if the action was actually send, a False indicating
            that the bot reached its configured timeout prior to preforming an action
        """
        await self.wait_for_my_turn()
        if not self.is_my_turn():
            return False

        return await self.send_action(api.GamePlayerIntent.PlayerIntentAllIn, {})

    async def raise_bet(self, raise_to: int):
        """sends action to raise bet after waiting for the bots turn

        Returns:
            a flag indicating True if the action was actually send, a False indicating
            that the bot reached its configured timeout prior to preforming an action

        Raises:
            ApiException: if the raise action was not valid given the state of the game
        """
        await self.wait_for_my_turn()
        if not self.is_my_turn():
            return False

        current_bet = help.convert_stack(self._player.current_bet)
        current_bet = sum(s.denomination * s.count for s in current_bet)
        amount = raise_to - current_bet

        result = await self._compute_valid_bet(amount)
        return await self.send_action(api.GamePlayerIntent.PlayerIntentRaise, result)

    async def ante(self):
        """sends action to raise bet after waiting for the bots turn

        Returns:
            a flag indicating True if the action was actually send, a False indicating
            that the bot reached its configured timeout prior to preforming an action

        Raises:
            ApiException: if the raise action was not valid given the state of the game
        """
        await self.wait_for_my_turn()
        if not self.is_my_turn():
            return False

        amount = self._current_state.table.current_round.bet
        if not amount:
            raise ex.CustomException("erm")

        result = await self._compute_valid_bet(amount)
        return await self.send_action(api.GamePlayerIntent.PlayerIntentAnte, result)

    async def call(self):
        """sends action to call after waiting for the bots turn

        Returns:
            a flag indicating True if the action was actually send, a False indicating
            that the bot reached its configured timeout prior to preforming an action

        Raises:
            ApiException: if the call action was not valid given the state of the game
        """
        await self.wait_for_my_turn()
        if not self.is_my_turn():
            return False

        amount = self._current_state.table.current_round.bet
        if not amount:
            raise ex.CustomException("erm")

        stack = help.convert_stack(self._player.current_bet)
        current_bet = sum(s.denomination * s.count for s in stack)
        amount -= current_bet

        result = await self._compute_valid_bet(amount)
        return await self.send_action(api.GamePlayerIntent.PlayerIntentCall, result)

    async def fold(self):
        """sends action to call after waiting for the bots turn

        Returns:
            a flag indicating True if the action was actually send, a False indicating
            that the bot reached its configured timeout prior to preforming an action

        Raises:
            ApiException: if the call action was not valid given the state of the game
        """
        await self.wait_for_my_turn()
        if not self.is_my_turn():
            return False

        return await self.send_action(api.GamePlayerIntent.PlayerIntentFold)

    async def send_action(self, intent: GamePlayerIntent | str, bet: dict[str, int] | None = None):
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

        player_intent: GamePlayerIntent | None = GamePlayerIntent.PlayerIntentUnset
        if isinstance(intent, str):
            try:
                player_intent = GamePlayerIntent(intent)
            except:
                pass

        dto = api.GamePlayerActionDTO(chips=bet, intent=player_intent)
        # TODO return if this is successful, bubble up to return all helpers too
        response: JustResponseMessageAny = await self._game_api.game_game_id_action_post(self._game_id, dto)
        return response.type != "error"

    async def wait_for_my_turn(self):
        """blocks until a game state is injected that signals it is this players turn

        Raises:
            ex.CostumException: if the bot is not joined to a game this will throw.
        """
        if not self._joined:
            raise ex.CustomException("you are not in the game - it can never be your turn")

        if not self.events.running:
            await self.start_events()

        if self._timeout >= 0:
            timeout = self._timeout
            while timeout > 0 and not self.is_my_turn():
                await asyncio.sleep(0.5)
                timeout -= 0.5
        else:
            while not self.is_my_turn():
                await asyncio.sleep(0.5)

    def is_my_turn(self) -> bool:
        if self._player is None:
            return False

        current_player_position = self._current_position()
        if current_player_position is None:
            return False

        return current_player_position == self._player.position

    def _break_chip(self, value: int, denominations: list[int], need_off_denom: bool = True) -> list[help.Chips]:
        # off_denom affects if we break into necessary and unnecessary chips
        pool = [d for d in sorted(denominations, reverse=True) if d < value]
        if not need_off_denom:
            step = min(denominations)
            on_denom = [d for d in pool if not d % step]
            pool = on_denom or pool

        result: list[help.Chips] = []
        remaining = value

        for denomination in pool:
            count = remaining // denomination
            if count > 0:
                result.append(help.Chips(denomination, count))
                remaining -= count * denomination
            if remaining == 0:
                break

        if remaining != 0:
            logger.error(f"{value} can not be subdivided with provided denominations {denominations}")

            raise ex.CustomException("invalid chip exchange request with current denominations")

        return result

    def _ingest_game_dto(self, state: api.GameGameDTO):
        """updates the internal model based on a new state

        alters the _current_state field as well as the _player and _current_stack
        """
        if state is None or state.table is None or state.table.players is None:
            logger.warning("state is none in _ingest_game_dto")
            return

        self._current_state = state

        for player in state.table.players:
            if player.user_id == self._user_id:
                self._joined = True
                self._player = player
                self._current_stack = sorted(
                    help.convert_stack(self._player.stack),
                    key=lambda i: i.denomination,
                    reverse=True,
                )
                break

        self._current_state = state

    async def _compute_valid_bet(self, amount: int) -> dict[str, int]:
        """computes a valid bet from a set of available chips and denominations to satisfy a required amount

        Args:
            amount: the amount we need to bet

        Constraints:
            - d > 0 ∀ d in denominations
            - k ∈ denominations ∀ k ∈ chips.keys()
            - k % d == 0 ∀ k and d such that d < k ∈ denominations
            - amount % d == 0 ∃ d ∈ denominations

        Returns:
            a mapping of string to integer expressing the bet that the player
            can make to satisfy the provided amount.
        """
        denominations = self._current_state.game_config.chip_denominations
        chips = {int(d): c for d, c in self._player.stack.items()}
        chips_sum = sum((d * c for d, c in chips.items()))
        if chips_sum < amount:
            # we are all in here
            return {str(d): c for d, c in chips.items()}

        denominations = sorted(denominations, reverse=True)
        valid_bet: dict[int, int] = {}
        for denomination in denominations:
            if denomination > amount:
                # this denomination is not useful for constructing the required set
                continue

            take = amount // denomination
            amount -= take * denomination
            valid_bet[denomination] = take

            if amount == 0:
                break

        if valid_bet is None:
            return None

        missing_chips: dict[int, int] = {}
        for denomination, count in valid_bet.items():
            if denomination not in chips:
                missing_chips[denomination] = count
                continue

            available = chips[denomination]
            if available < count:
                # if we do not have enough chips to cover this denomination
                # we zero out our count and track the reminaing chips required
                missing_chips[denomination] = count - available
                chips[denomination] = 0
            else:
                # otherwise we know that we can cover this denomination, just do that
                chips[denomination] -= count

        give: dict[int, int] = {d: 0 for d in denominations}
        receive: dict[int, int] = {d: 0 for d in denominations}
        # here we just need to do any required exchanges from our remaining
        # chips in order to ensure that we cover this bet
        for denomination, count in sorted(missing_chips.items()):
            # we need to get count of denomination, end of story.
            # this iteration of the loop cannot terminate without satisfying
            # this requirement
            if count == 0:
                break

            for d, c in sorted(chips.items(), reverse=True):
                if count <= 0:
                    break

                if c == 0:
                    continue

                # we are chipping down with this part of the exchange
                if d > denomination:
                    exchange_rate = d // denomination
                    while chips[d] > 0 and count > 0:
                        chips[d] -= 1
                        give[d] += 1
                        receive[denomination] += exchange_rate
                        count -= exchange_rate

                # we are chipping down with this part of the exchange
                elif d < denomination:
                    exchange_rate = denomination // d
                    while chips[d] >= exchange_rate and count > 0:
                        chips[d] -= exchange_rate
                        give[d] += exchange_rate
                        receive[denomination] += 1
                        count -= 1

        give_some = sum((d * c for d, c in give.items()))

        if give_some > 0:
            try:
                await self.exchange_chips(
                    [help.Chips(d, c) for d, c in give.items()],
                    [help.Chips(d, c) for d, c in receive.items()],
                )

            except Exception:
                print("encountered error exchanging chips")
                print(f"receive: {receive}")
                print(f"give: {give}")
                print(f"bet: {valid_bet}")
                print(f"stack: {self._player.stack}")
                raise

        return {str(d): c for d, c in valid_bet.items()}

    def _get_valid_bet(self, denominations: list[int], amount: int) -> dict[int, int] | None:
        """gets a valid bet which satisfies amount from denominations"""

    def _held(self) -> dict[int, int]:
        held: dict[int, int] = {}
        for s in self._current_stack:
            held[s.denomination] = held.get(s.denomination, 0) + s.count
        return held

    def _current_position(self) -> int | None:
        if self._current_state is None or self._current_state.table is None:
            return None
        if self._current_state.table.current_round is None:
            return None
        return self._current_state.table.current_round.current_player_position

    def _ingest_websocket_event(self, event: WebSocketEvent):
        if event.data is not None and isinstance(event.data, api.GameGameDTO):
            self._ingest_game_dto(event.data)

    async def __aenter__(self) -> "PokerBot":
        # we call this to ensure that the event hub as been started prior
        # to exposing the bot within an async with block. this follows a
        # pattern for starting/stopping an async context.
        await self.start_events()
        return self

    async def __aexit__(self, exc_type, exc, tb) -> None:
        if self._state_subscription:
            self._state_subscription.unsubscribe()
        await self.stop_events()
