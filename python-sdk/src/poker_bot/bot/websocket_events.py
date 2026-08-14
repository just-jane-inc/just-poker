import asyncio
import inspect
import json
import logging
from collections.abc import AsyncGenerator, AsyncIterator, Awaitable, Callable
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Union, get_args, get_origin
from urllib.parse import urlsplit, urlunsplit

import websockets

import poker_bot.bot.poker_exceptions as ex
from openapi_client import GameGameDTO, GamePlayerActionDTO, GamePlayerDTO, GameRoundDTO

logger = logging.getLogger("websocket")

GameStateCallback = Callable[[GameGameDTO], Awaitable[None] | None]
EventCallback = Callable[["WebSocketEvent"], Awaitable[None] | None]
CloseCallback = Callable[["CloseInfo"], Awaitable[None] | None]


@dataclass(frozen=True)
class CloseInfo:
    code: int | None
    reason: str


def deserialize(data: Any, data_class: Any) -> Any:
    if data is None:
        return None

    if data_class is dict:
        return data

    origin = get_origin(data_class)

    if origin is Union:
        # Handles if enum dataclass uses Optional
        for subtype in get_args(data_class):
            if subtype is type(None):
                continue
            try:
                return deserialize(data, subtype)
            except (TypeError, ValueError):
                continue

    if origin is list:
        subtype = get_args(data_class)[0]
        return [deserialize(item, subtype) for item in data]

    if hasattr(data_class, "from_dict"):
        return data_class.from_dict(data)

    return data


class UnknownWebSocketEventType:
    """Unknown event holder to retain original string type"""

    def __init__(self, value: str):
        self.value = value
        self.dto_type = dict
        self.name = "UNKNOWN"

    def __repr__(self):
        return f"UnknownWebSocketEventType({self.value!r})"


class WebSocketEventType(Enum):
    """Known Websocket Events with their associated DTO if available"""

    def __new__(cls, value: str, obj_type):
        obj = object.__new__(cls)
        obj._value_ = value
        obj.obj_type = obj_type
        return obj

    UNKNOWN = "", Any
    WELCOME = "welcome", GameGameDTO
    GAME_STATE_UPDATE = "game_state_update", GameGameDTO
    PLAYER_ACTION = "player_action", GamePlayerActionDTO
    PAYOUT = "payout", dict
    ROUND_START = "round_start", GameRoundDTO
    HAND_STARTED = "hand_started", dict
    GAME_OVER = "game_over", list[GamePlayerDTO]
    STARTING_GAME = "starting_game", GameGameDTO

    @property
    def data_class(self):
        return self.obj_type

    @classmethod
    def from_str(
        cls, event_type: str
    ) -> "WebSocketEventType | UnknownWebSocketEventType":
        try:
            return cls(event_type)
        except ValueError:
            return UnknownWebSocketEventType(event_type)

    def __str__(self) -> str:
        if self._value_ == "":
            return "Unknown"
        return self._value_


@dataclass(frozen=True)
class WebSocketEvent:
    event_type: WebSocketEventType | UnknownWebSocketEventType
    _data: dict | list | None
    id: int = 0
    time_sent: str = ""
    raw: dict[str, Any] = field(default_factory=dict)

    @property
    def data(self) -> Any:
        if self._data is None:
            return self._data

        if not isinstance(self.event_type, WebSocketEventType):
            return self._data

        data_class = self.event_type.data_class
        if data_class is dict:
            return self._data

        try:
            return deserialize(self._data, self.event_type.data_class)
        except Exception as e:
            logger.warning(f"could not parse event data: {e}")
            return self._data


def build_state_ws_url(base_url: str, game_id: str) -> str:
    if not base_url:
        raise ex.CustomException("base url not provided")

    if not game_id:
        raise ex.CustomException("game id not provided")

    parts = urlsplit(base_url)
    scheme = {"http": "ws", "https": "wss"}.get(parts.scheme, parts.scheme)
    path = f"{parts.path.rstrip('/')}/game/{game_id}/state/ws"
    return urlunsplit((scheme, parts.netloc, path, "", ""))


def parse_event(message: str | bytes) -> WebSocketEvent:
    """Parse websocket event data"""
    if isinstance(message, (bytes, bytearray)):
        message = message.decode("utf-8")

    try:
        payload = json.loads(message)
    except json.JSONDecodeError as e:
        raise ex.CustomException(f"received non json message on update feed: {e}")

    if not isinstance(payload, dict):
        raise ex.CustomException("received non object message on update feed")

    if recv_data := payload.get("data", {}):
        _event_type = payload.get("event_type", "")

        event_type = WebSocketEventType.from_str(_event_type)

        logger.info(f"received event [{_event_type}]: {json.dumps(recv_data)}")

        time_sent = payload.get("time_sent", "")
        try:
            event_id = int(payload.get("id", "0"))
        except (TypeError, ValueError):
            raise ex.CustomException(f"non integer event id: {payload.get('id')!r}")
    else:
        raise ex.CustomException(f"unexpected data received: {payload}")

    if event_type == WebSocketEventType.UNKNOWN or isinstance(
        event_type, UnknownWebSocketEventType
    ):
        logger.debug(f"unknown event type from websocket: {_event_type}")

    state = None
    if isinstance(recv_data, dict) or isinstance(recv_data, list):
        state = recv_data

    return WebSocketEvent(
        event_type=event_type,
        _data=state,
        id=event_id,
        time_sent=str(time_sent),
        raw=payload,
    )


class WebSocketStream:
    """Stream of WebSocket Events. Blocking"""

    def __init__(
        self,
        base_url: str,
        token: str,
        game_id: str,
        *,
        reconnect: bool = True,
        max_retries: int = 0,
        retry_backoff: float = 1.0,
        max_retry_backoff: float = 30.0,
        on_state: EventCallback | None = None,
        on_game_state: GameStateCallback | None = None,
        on_close: CloseCallback | None = None,
    ):
        if not token:
            raise ex.CustomException("api token not provided")

        self._url = build_state_ws_url(base_url, game_id)
        self._token = token
        self._game_id = game_id
        self._reconnect = reconnect
        self._max_retries = max_retries
        self._retry_backoff = retry_backoff
        self._max_retry_backoff = max_retry_backoff
        self._on_state = on_state
        self._on_game_state = on_game_state
        self._conn: websockets.ClientConnection | None = None
        self._closed = False
        self._on_close = on_close

    @property
    def url(self) -> str:
        return self._url

    async def connect(self) -> "WebSocketStream":
        if self._conn is not None:
            return self

        self._closed = False
        self._conn = await websockets.connect(
            self._url, additional_headers={"Authorization": f"Bearer {self._token}"}
        )

        logger.debug(f"connected to game state feed for game {self._game_id}")
        return self

    async def _fire_close_event(self, *, code: int | None, reason: str) -> None:
        if self._on_close is None:
            return
        info = CloseInfo(code=code, reason=reason)
        try:
            await _call(self._on_close, info)
        except asyncio.CancelledError:
            raise
        except Exception as e:
            logger.error(f"error in on_close handler: {e}")

    async def close(self):
        if self._closed and self._conn and self._conn.state.OPEN:
            await self._fire_close_event(code=-1, reason="Manual close requested")
        self._closed = True
        if self._conn is not None:
            await self._conn.close()
            self._conn = None

    async def __aenter__(self) -> "WebSocketStream":
        return await self.connect()

    async def __aexit__(self, exc_type, exc, tb):
        await self.close()

    async def events(self) -> AsyncGenerator[WebSocketEvent, None]:
        attempts = 0
        while not self._closed:
            try:
                await self.connect()
                attempts = 0
                assert self._conn is not None

                async for message in self._conn:
                    try:
                        event = parse_event(message)
                    except ex.CustomException as e:
                        logger.error(f"dropping bad update message: {e}")
                        continue

                    if not isinstance(event.event_type, WebSocketEventType):
                        continue

                    if event is None:
                        continue

                    if event is not None and self._on_state is not None:
                        try:
                            await _call(self._on_state, event)
                        except asyncio.CancelledError:
                            raise
                        except Exception as e:
                            logger.error(f"error in on_state hook: {e}", exc_info=True)

                    if (
                        event is not None
                        and event.data is not None
                        and self._on_game_state is not None
                        and isinstance(event.data, GameGameDTO)
                    ):
                        try:
                            await _call(self._on_game_state, event.data)
                        except asyncio.CancelledError:
                            raise
                        except Exception as e:
                            logger.error(
                                f"error in on_game_state hook: {e}", exc_info=True
                            )

                    if event is not None:
                        yield event

                    if (
                        event is not None
                        and event.event_type == WebSocketEventType.GAME_OVER
                    ):
                        # We do not get a 404 on connect retry attempts on ended games, so we have to
                        self._reconnect = False
                        await asyncio.sleep(0.5)
                        await self.close()
                        return  # Closes generator

                if self._conn is not None:
                    await self._fire_close_event(
                        code=self._conn.close_code, reason=self._conn.close_reason or ""
                    )

            except asyncio.CancelledError:
                raise

            except (
                websockets.exceptions.ConnectionClosedError,
                websockets.exceptions.ConnectionClosedOK,
            ) as e:
                close = e.rcvd or e.sent
                if close is not None:
                    await self._fire_close_event(code=close.code, reason=close.reason)

                if e.rcvd is not None and e.rcvd.code in [
                    websockets.CloseCode.NORMAL_CLOSURE,
                    websockets.CloseCode.BAD_GATEWAY,
                    websockets.CloseCode.GOING_AWAY,
                    websockets.CloseCode.NO_STATUS_RCVD,
                ]:
                    logger.info(f"received normal close event: {e}")
                    self._reconnect = False
                    await self.close()

                else:
                    logger.warning(f"received abnormal close event: {e}")
                    self._reconnect = False
                    await self.close()

            except Exception as e:
                if self._closed:
                    return

                await self.close()

                if not self._reconnect:
                    raise ex.CustomException(f"game state feed failed: {e}")

                logger.warning(f"game state feed dropped, reconnecting: {e}")
            else:
                self._conn = None

            if self._closed or not self._reconnect:
                return

            attempts += 1
            if self._max_retries and attempts > self._max_retries:
                raise ex.CustomException(
                    f"game state feed gave up after {self._max_retries} retries"
                )

            delay = min(
                self._retry_backoff * (2 ** (attempts - 1)), self._max_retry_backoff
            )
            await asyncio.sleep(delay)

    async def states(self) -> AsyncGenerator[WebSocketEvent, None]:
        async for event in self.events():
            if event is not None and event.data is not None:
                yield event

    def __aiter__(self) -> AsyncIterator[WebSocketEvent]:
        return self.states()

    async def next_state(self, timeout: float | None = None) -> WebSocketEvent:
        states = self.states()

        async def _next() -> WebSocketEvent:
            try:
                return await states.__anext__()
            except StopAsyncIteration:
                raise ex.CustomException("update feed closed before a state arrived")

        try:
            return await asyncio.wait_for(_next(), timeout)
        finally:
            await states.aclose()


class WebSocketListener:
    """Listener for Updates. Non-blocking, assign a callback function to run whenever an update is received"""

    def __init__(self, base_url: str, token: str, game_id: str, **kwargs):
        self._stream = WebSocketStream(base_url, token, game_id, **kwargs)
        self._game_state_handlers: list[GameStateCallback] = []
        self._event_handlers: dict[str, list[EventCallback]] = {}
        self._task: asyncio.Task | None = None

    @property
    def stream(self) -> WebSocketStream:
        return self._stream

    @property
    def running(self) -> bool:
        return self._task is not None and not self._task.done()

    def on_game_state(self, callback: GameStateCallback) -> GameStateCallback:
        self._game_state_handlers.append(callback)
        return callback

    def on_event(
        self, *event_types: "WebSocketEventType | str"
    ) -> Callable[[EventCallback], EventCallback]:
        """Register this handler for one or more event types.

        Ex.
        @listener.on_event(WebSocketEventType.WELCOME,
                           WebSocketEventType.GAME_STATE_UPDATE)
        # @listener.on_event("*") # to subscribe to all events
        async def on_update(event: WebSocketEvent):
        """
        if not event_types:
            raise ex.CustomException("on_event needs at least one event type")

        keys = [str(event_type) for event_type in event_types]

        def decorator(callback: EventCallback) -> EventCallback:
            for key in keys:
                self._event_handlers.setdefault(key, []).append(callback)
            return callback

        return decorator

    async def start(self):
        if self.running:
            return

        await self._stream.connect()
        self._task = asyncio.create_task(self._run())

    async def stop(self):
        await self._stream.close()
        if self._task is not None:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
            self._task = None

    async def run_forever(self):
        await self._run()

    async def __aenter__(self) -> "WebSocketListener":
        await self.start()
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.stop()

    async def _run(self):
        async for event in self._stream.events():
            handlers: list[tuple[EventCallback | GameStateCallback, Any]] = []

            # Send to listeners subscribed to *all* events
            for handler in self._event_handlers.get("*", []):
                handlers.append((handler, event))

            for handler in self._event_handlers.get(str(event.event_type), []):
                handlers.append((handler, event))

            if event.data is not None and isinstance(event.data, GameGameDTO):
                for handler in self._game_state_handlers:
                    handlers.append((handler, event.data))

            for handler, arg in handlers:
                try:
                    await _call(handler, arg)
                except asyncio.CancelledError:
                    raise
                except Exception as e:
                    logger.error(
                        f"error in update handler {handler!r}: {e}", exc_info=True
                    )


async def _call(callback, arg):
    result = callback(arg)
    if inspect.isawaitable(result):
        await result
