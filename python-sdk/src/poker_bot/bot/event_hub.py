import asyncio
import logging
from collections.abc import (
    AsyncGenerator,
    AsyncIterator,
    Awaitable,
    Callable,
    Iterable,
)
from typing import Any, Protocol, Union, runtime_checkable

import poker_bot.bot.poker_exceptions as ex
from poker_bot.bot.websocket_events import (
    CloseInfo,
    WebSocketEvent,
    WebSocketEventType,
    WebSocketStream,
    _call,
)

logger = logging.getLogger("event_hub")

Event = WebSocketEvent
EventType = WebSocketEventType

EventHandler = Callable[[Event], Awaitable[None] | None]
ALL_EVENTS = "*"


@runtime_checkable
class EventTransport(Protocol):
    """Abstract setup"""

    async def connect(self) -> Any: ...
    async def close(self) -> None: ...
    def events(self) -> AsyncGenerator[Event, None]: ...


class EventSubscriber:
    # __slots__ in python is basically the "allowed" attributes, rather than blanket whatever is its __dict__
    # aka, this prevents arbitrarily adding new attribute stuff to it, which is cool cause now it's limited
    __slots__ = ("__weakref__", "_active", "_callback", "_hub", "_keys")

    def __init__(self, hub: "EventHub", keys: tuple[str, ...], callback: EventHandler):
        self._hub: EventHub | None = hub
        self._keys = keys
        self._callback = callback
        self._active = True

    @property
    def active(self) -> bool:
        return self._active

    @property
    def event_types(self) -> tuple[str, ...]:
        return self._keys

    @property
    def callback(self) -> EventHandler:
        return self._callback

    def unsubscribe(self) -> bool:
        if not self._active:
            return False

        self._active = False

        # detatch
        hub, self._hub = self._hub, None
        if hub is not None:
            hub._remove(self)
        return True

    def __enter__(self) -> "EventSubscriber":
        return self

    def __exit__(self, exc_type, exc, tb) -> bool:
        self.unsubscribe()
        return False

    async def __aenter__(self) -> "EventSubscriber":
        return self

    async def __aexit__(self, exc_type, exc, tb) -> bool:
        self.unsubscribe()
        return False

    def __del__(self):
        try:
            self.unsubscribe()
        except Exception:
            pass

    def __repr__(self) -> str:
        state = "active" if self._active else "unsubscribed"
        return f"EventSubscriber({'|'.join(self._keys)}, {state})"


def _normalize(event_type: "EventTypeArg") -> tuple[str, ...]:
    if isinstance(event_type, str) or isinstance(event_type, WebSocketEventType):
        return (str(event_type),)

    if isinstance(event_type, Iterable):
        keys = tuple(str(t) for t in event_type)
        if not keys:
            raise ex.CustomException("subscribe needs at least one event type")
        return keys

    raise ex.CustomException(f"unsupported event type: {event_type!r}")


EventTypeArg = Union[str, WebSocketEventType, Iterable[str | WebSocketEventType]]


class EventHub:
    def __init__(self, transport: EventTransport, *, name: str = "event-hub"):
        self._transport = transport
        self._name = name
        self._subscribers: dict[str, list[EventSubscriber]] = {}
        self._task: asyncio.Task | None = None
        self._closed = asyncio.Event()
        self._close_info: CloseInfo | None = None

    @property
    def running(self) -> bool:
        return self._task is not None and not self._task.done()

    @property
    def close_info(self) -> CloseInfo | None:
        return self._close_info

    @property
    def transport(self) -> EventTransport:
        return self._transport

    def subscribe(self, event_type: EventTypeArg, callback: EventHandler) -> EventSubscriber:
        keys = _normalize(event_type)
        subscriber = EventSubscriber(self, keys, callback)
        for key in keys:
            self._subscribers.setdefault(key, []).append(subscriber)
        return subscriber

    def on_event(self, *event_types: str | WebSocketEventType) -> Callable[[EventHandler], EventHandler]:
        if not event_types:
            raise ex.CustomException("on_event needs at least one event type, or '*' for all")

        def decorator(callback: EventHandler) -> EventHandler:
            subscriber = self.subscribe(event_types, callback)
            subscriptions = getattr(callback, "event_subscriptions", None)
            if subscriptions is None:
                subscriptions = []
                try:
                    callback.event_subscriptions = subscriptions
                    callback.unsubscribe = lambda: any([s.unsubscribe() for s in subscriptions])
                except AttributeError:
                    logger.debug(f"could not attach subscription to {callback!r}")

            subscriptions.append(subscriber)
            return callback

        return decorator

    def _remove(self, subscriber: EventSubscriber) -> None:
        for key in subscriber.event_types:
            handlers = self._subscribers.get(key)
            if not handlers:
                continue
            try:
                handlers.remove(subscriber)
            except ValueError:
                continue
            if not handlers:
                self._subscribers.pop(key, None)

    def subscriber_count(self, event_type: EventTypeArg | None = None) -> int:
        if event_type is None:
            return sum(len(v) for v in self._subscribers.values())
        return sum(len(self._subscribers.get(k, ())) for k in _normalize(event_type))

    async def wait_for(
        self,
        event_type: EventTypeArg,
        timeout: float | None = None,
        predicate: Callable[[Event], bool] | None = None,
    ) -> Event:
        # Asyncio and locks are weird, this seems to be the pattern that is preferred
        loop = asyncio.get_running_loop()
        future: asyncio.Future[Event] = loop.create_future()

        def _on_event(event: Event) -> None:
            if future.done():
                return
            if predicate is not None and not predicate(event):
                return
            future.set_result(event)

        with self.subscribe(event_type, _on_event):
            await self.start()
            return await asyncio.wait_for(future, timeout)

    async def wait_closed(self, timeout: float | None = None) -> CloseInfo | None:
        await asyncio.wait_for(self._closed.wait(), timeout)
        return self._close_info

    async def stream(self, event_type: EventTypeArg = ALL_EVENTS, *, maxsize: int = 0) -> AsyncGenerator[Event, None]:
        """Stream events directly"""
        queue: asyncio.Queue[Event] = asyncio.Queue(maxsize=maxsize)

        def _enqueue(event: Event) -> None:
            try:
                queue.put_nowait(event)
            except asyncio.QueueFull:
                logger.warning("stream consumer is behind, dropping event")

        with self.subscribe(event_type, _enqueue):
            await self.start()
            closed = asyncio.ensure_future(self._closed.wait())
            getter: asyncio.Future | None = None
            try:
                while True:
                    getter = asyncio.ensure_future(queue.get())
                    await asyncio.wait((getter, closed), return_when=asyncio.FIRST_COMPLETED)

                    if getter.done():
                        yield getter.result()
                        continue

                    getter.cancel()
                    while not queue.empty():
                        yield queue.get_nowait()
                    return
            finally:
                closed.cancel()
                if getter is not None:
                    getter.cancel()

    def __aiter__(self) -> AsyncIterator[Event]:
        """Allows 'async for x in hub' syntax on an EventHub"""
        return self.stream()

    async def start(self) -> "EventHub":
        if self.running:
            return self

        self._closed.clear()
        await self._transport.connect()
        self._task = asyncio.create_task(self._run(), name=self._name)
        return self

    async def stop(self) -> None:
        await self._transport.close()

        task, self._task = self._task, None
        if task is not None:
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass

        self._closed.set()

    async def run_forever(self) -> None:
        await self._run()

    async def __aenter__(self) -> "EventHub":
        return await self.start()

    async def __aexit__(self, exc_type, exc, tb) -> None:
        await self.stop()

    async def _run(self) -> None:
        try:
            async for event in self._transport.events():
                await self._dispatch(event)
        except asyncio.CancelledError:
            raise
        except Exception as e:
            logger.error(f"event feed failed: {e}", exc_info=True)
            self._close_info = CloseInfo(code=None, reason=str(e))
            raise
        finally:
            self._closed.set()

    async def _dispatch(self, event: Event) -> None:
        targets = list(self._subscribers.get(ALL_EVENTS, ()))
        targets += list(self._subscribers.get(str(event.event_type), ()))

        for subscriber in targets:
            if not subscriber.active:
                continue
            try:
                await _call(subscriber.callback, event)
            except asyncio.CancelledError:
                raise
            except Exception as e:
                logger.error(f"error in handler {subscriber!r}: {e}", exc_info=True)

    def _note_close(self, info: CloseInfo) -> None:
        self._close_info = info


def websocket_event_hub(base_url: str, token: str, game_id: str, **kwargs) -> EventHub:
    hub_holder: dict[str, EventHub] = {}

    def _on_close(info: CloseInfo) -> None:
        hub = hub_holder.get("hub")
        if hub is not None:
            hub._note_close(info)

    kwargs.setdefault("on_close", _on_close)
    stream = WebSocketStream(base_url, token, game_id, **kwargs)
    hub = EventHub(stream, name=f"event-hub[{game_id}]")
    hub_holder["hub"] = hub
    return hub
