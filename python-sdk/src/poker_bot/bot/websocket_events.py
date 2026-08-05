import asyncio
import inspect
import json
import logging

from dataclasses import dataclass, field
from typing import Any, AsyncGenerator, AsyncIterator, Awaitable, Callable
from urllib.parse import urlsplit, urlunsplit

import websockets

import poker_bot.bot.poker_exceptions as ex
from openapi_client import GameGameDTO

logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger("websocket")

GameStateCallback = Callable[[GameGameDTO], Awaitable[None] | None]
EventCallback = Callable[["WebSocketEvent"], Awaitable[None] | None]

# Note: this is the only state now
GAME_STATE_EVENT = "game_state_update"
VALID_STATES = [ GAME_STATE_EVENT ]

@dataclass(frozen=True)
class WebSocketEvent:
    event_type: str
    data: Any | GameGameDTO | None # only game dto for now
    id: int = 0
    time_sent: str = ""
    raw: dict[str, Any] = field(default_factory=dict)


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
    if isinstance(message, (bytes, bytearray)):
        message = message.decode("utf-8")

    try:
        payload = json.loads(message)
    except json.JSONDecodeError as e:
        raise ex.CustomException(f"received non json message on update feed: {e}")

    if not isinstance(payload, dict):
        raise ex.CustomException("received non object message on update feed")

    if recv_data := payload.get("data", {}):
        event_type = payload.get("event_type", "")

        if event_type == "welcome": # Note: Only value as it is the only ws message at the moment. On join response
            event_type = GAME_STATE_EVENT

        time_sent = payload.get("time_sent", "")
        try:
            event_id = int(payload.get("id", "0"))
        except (TypeError, ValueError):
            raise ex.CustomException(f"non integer event id: {payload.get('id')!r}")
    else:
        raise ex.CustomException(f"unexpected data received: {payload}")

    if event_type not in VALID_STATES:
        raise ex.CustomException(f"could not find valid model for event type: {event_type}")

    state = None
    if isinstance(recv_data, dict):
        try:
            if event_type == GAME_STATE_EVENT: # Note: Only game dto for now
                state = GameGameDTO.from_dict(recv_data)
        except Exception as e:
            raise ex.CustomException(f"could not parse game state from update: {e}")

    return WebSocketEvent(
        event_type=event_type,
        data=state,
        id=event_id,
        time_sent=str(time_sent),
        raw=payload,
    )


class GameStateStream:
    def __init__(self, base_url: str, token: str, game_id: str,
        *,
        reconnect: bool = True,
        max_retries: int = 0,
        retry_backoff: float = 1.0,
        max_retry_backoff: float = 30.0,
        on_state: GameStateCallback | None = None,
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
        self._conn = None
        self._closed = False

    @property
    def url(self) -> str:
        return self._url

    async def connect(self) -> "GameStateStream":
        if self._conn is not None:
            return self

        self._closed = False
        self._conn = await websockets.connect(self._url,
            additional_headers={"Authorization": f"Bearer {self._token}"})

        logger.debug(f"connected to game state feed for game {self._game_id}")
        return self

    async def close(self):
        self._closed = True
        if self._conn is not None:
            await self._conn.close()
            self._conn = None

    async def __aenter__(self) -> "GameStateStream":
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

                    if event.event_type != GAME_STATE_EVENT:
                        continue

                    if event.data is not None and self._on_state is not None:
                        try:
                            await _call(self._on_state, event.data)
                        except asyncio.CancelledError:
                            raise
                        except Exception as e:
                            logger.error(f"error in on_state hook: {e}", exc_info=True)
                    yield event

            except asyncio.CancelledError:
                raise
            except Exception as e:
                self._conn = None
                if self._closed:
                    return

                if not self._reconnect:
                    raise ex.CustomException(f"game state feed failed: {e}")

                logger.warning(f"game state feed dropped, reconnecting: {e}")
            else:
                self._conn = None

            if self._closed or not self._reconnect:
                return

            attempts += 1
            if self._max_retries and attempts > self._max_retries:
                raise ex.CustomException(f"game state feed gave up after {self._max_retries} retries")

            delay = min(self._retry_backoff * (2 ** (attempts - 1)),
                        self._max_retry_backoff)
            await asyncio.sleep(delay)

    async def states(self) -> AsyncGenerator[GameGameDTO, None]:
        async for event in self.events():
            if event.data is not None:
                yield event.data

    def __aiter__(self) -> AsyncIterator[GameGameDTO]:
        return self.states()

    async def next_state(self, timeout: float | None = None) -> GameGameDTO:
        states = self.states()

        async def _next() -> GameGameDTO:
            try:
                return await states.__anext__()
            except StopAsyncIteration:
                raise ex.CustomException("update feed closed before a state arrived")

        try:
            return await asyncio.wait_for(_next(), timeout)
        finally:
            await states.aclose()


class GameStateListener:
    def __init__(self, base_url: str, token: str, game_id: str, **kwargs):
        self._stream = GameStateStream(base_url, token, game_id, **kwargs)
        self._state_handlers: list[GameStateCallback] = []
        self._event_handlers: dict[str, list[EventCallback]] = {}
        self._task: asyncio.Task | None = None

    @property
    def stream(self) -> GameStateStream:
        return self._stream

    @property
    def running(self) -> bool:
        return self._task is not None and not self._task.done()

    def on_state(self, callback: GameStateCallback) -> GameStateCallback:
        self._state_handlers.append(callback)
        return callback

    def on_event(self, event_type: str) -> Callable[[EventCallback], EventCallback]:
        def decorator(callback: EventCallback) -> EventCallback:
            self._event_handlers.setdefault(event_type, []).append(callback)
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

    async def __aenter__(self) -> "GameStateListener":
        await self.start()
        return self

    async def __aexit__(self, exc_type, exc, tb):
        await self.stop()

    async def _run(self):
        async for event in self._stream.events():
            handlers: list[tuple[EventCallback | GameStateCallback, Any]] = []

            for handler in self._event_handlers.get("*", []):
                handlers.append((handler, event))

            for handler in self._event_handlers.get(event.event_type, []):
                handlers.append((handler, event))

            if event.data is not None:
                for handler in self._state_handlers:
                    handlers.append((handler, event.data))

            for handler, arg in handlers:
                try:
                    await _call(handler, arg)
                except asyncio.CancelledError:
                    raise
                except Exception as e:
                    logger.error(f"error in update handler {handler!r}: {e}", exc_info=True)


async def _call(callback, arg):
    result = callback(arg)
    if inspect.isawaitable(result):
        await result
