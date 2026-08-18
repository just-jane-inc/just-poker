import logging

logging.basicConfig(
    filename="app.log",  # Name of the file
    filemode="a",  # 'a' to append, 'w' to overwrite each run
    format="%(asctime)s - %(levelname)s - [%(name)s] %(message)s",
    level=logging.INFO,  # Capture INFO, WARNING, ERROR, and CRITICAL
)

from poker_bot.event_hub import (
    ALL_EVENTS,
    Event,
    EventHandler,
    EventHub,
    EventSubscriber,
    EventTransport,
    EventType,
    websocket_event_hub,
)
from poker_bot.websocket_events import (
    CloseInfo,
    WebSocketEvent,
    WebSocketEventType,
    WebSocketListener,
    WebSocketStream,
)

__all__ = [
    "ALL_EVENTS",
    "CloseInfo",
    "Event",
    "EventHandler",
    "EventHub",
    "EventSubscriber",
    "EventTransport",
    "EventType",
    "WebSocketEvent",
    "WebSocketEventType",
    "WebSocketListener",
    "WebSocketStream",
    "websocket_event_hub",
]
