import asyncio
from collections.abc import AsyncGenerator

import pytest

from openapi_client import GamePayoutEventDTO
from poker_bot.event_hub import (
    ALL_EVENTS,
    Event,
    EventHub,
    EventType,
)

import poker_bot.poker_helpers as help
from poker_bot.websocket_events import WebSocketEvent

"""
Test behaviours of the eventhub
"""


def make_event(event_type: EventType, data: dict | list | None = None) -> Event:
    return WebSocketEvent(event_type=event_type, _data=data or {})


class FakeTransport:
    def __init__(self, events: list[Event], *, hold_open: bool = False):
        self._events = events
        self._hold_open = hold_open
        self.connected = False
        self.closed = False

    async def connect(self) -> "FakeTransport":
        self.connected = True
        return self

    async def close(self) -> None:
        self.closed = True

    async def events(self) -> AsyncGenerator[Event, None]:
        for event in self._events:
            yield event
            await asyncio.sleep(0)

        while self._hold_open:
            await asyncio.sleep(0.01)


@pytest.mark.asyncio
async def test_subscriber_receives_matching_events():
    received: list[Event] = []

    payout_event = GamePayoutEventDTO(player_id='45', chips={'10': 20})
    transport = FakeTransport(
        [
            make_event(EventType.WELCOME),
            make_event(EventType.PAYOUT, [payout_event.to_dict()]),
            make_event(EventType.GAME_ENDING),
        ]
    )
    hub = EventHub(transport)

    hub.subscribe(EventType.PAYOUT, lambda e: received.append(e))

    async with hub:
        await hub.wait_closed(timeout=5)

    assert len(received) == 1 and len(received[0].data) == 1
    assert received[0].data[0].to_dict() == {"chips": {'10': 20}, 'player_id': '45'}


@pytest.mark.asyncio
async def test_subscribe_accepts_multiple():
    received: list[Event] = []
    hub = EventHub(
        FakeTransport(
            [
                make_event(EventType.WELCOME),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PAYOUT),
            ]
        )
    )

    hub.subscribe((EventType.WELCOME, EventType.PAYOUT), received.append)

    async with hub:
        await hub.wait_closed(timeout=5)

    assert len(received) == 2


@pytest.mark.asyncio
async def test_wildcard_subscription():
    received: list[Event] = []
    hub = EventHub(FakeTransport([make_event(EventType.WELCOME), make_event(EventType.ROUND_START)]))

    hub.subscribe(ALL_EVENTS, received.append)

    async with hub:
        await hub.wait_closed(timeout=5)

    assert len(received) == 2


@pytest.mark.asyncio
async def test_block_unsubscribes_on_exit():
    hub = EventHub(FakeTransport([], hold_open=True))

    with hub.subscribe(EventType.GAME_ENDING, lambda e: None) as subscriber:
        assert subscriber.active
        assert hub.subscriber_count(EventType.GAME_ENDING) == 1

    assert not subscriber.active
    assert hub.subscriber_count(EventType.GAME_ENDING) == 0


@pytest.mark.asyncio
async def test_unsubscribe_safety():
    hub = EventHub(FakeTransport([]))

    subscriber = hub.subscribe(EventType.GAME_ENDING, lambda e: None)
    assert subscriber.unsubscribe() is True
    assert subscriber.unsubscribe() is False
    assert subscriber.unsubscribe() is False


@pytest.mark.asyncio
async def test_multi_unsubscribe():
    hub = EventHub(FakeTransport([]))

    subscriber = hub.subscribe((EventType.WELCOME, EventType.PAYOUT), lambda e: None)
    assert hub.subscriber_count() == 2

    subscriber.unsubscribe()
    assert hub.subscriber_count() == 0


@pytest.mark.asyncio
async def test_wait_for_event_then_unsub():
    hub = EventHub(
        FakeTransport(
            [
                make_event(EventType.WELCOME),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.GAME_STATE_UPDATE),
                make_event(EventType.PAYOUT, [{"chips": {'10': 20}, 'player_id': '45'}]),
            ],
            hold_open=True,
        )
    )

    try:
        event = await hub.wait_for(EventType.PAYOUT, timeout=5)
        assert event.data[0].to_dict() == {"chips": {'10': 20}, 'player_id': '45'}
        assert hub.subscriber_count() == 0
    finally:
        await hub.stop()


@pytest.mark.asyncio
async def test_wait_for_timeout():
    hub = EventHub(FakeTransport([], hold_open=True))

    try:
        with pytest.raises(asyncio.TimeoutError):
            await hub.wait_for(EventType.GAME_ENDING, timeout=0.1)

        assert hub.subscriber_count() == 0
    finally:
        await hub.stop()


@pytest.mark.asyncio
async def test_wait_for_fancy_predicate():

    hub = EventHub(
        FakeTransport(
            [
                make_event(EventType.PAYOUT, [{"chips": {'10': 20}, 'player_id': '45'}]),
                make_event(EventType.PAYOUT, [{"chips": {'10': 20, '100': 5}, 'player_id': '420'}]),
                make_event(EventType.PAYOUT, [{"chips": {'50': 1, '10': 1, '1': 7}, 'player_id': '67'}]),  # Is this considered a bribe
                make_event(EventType.PAYOUT, [{"chips": {'10': 5, '100': 5}, 'player_id': '68'}]),
                make_event(EventType.PAYOUT, [{"chips": {'10': 1, '500': 2}, 'player_id': '69'}, {"chips": {'10': 1, '500': 2}, 'player_id': '42'}]),
            ],
            hold_open=True,
        )
    )

    try:
        event = await hub.wait_for(EventType.PAYOUT, timeout=5, predicate=lambda e: help.chip_sum(e.data[0].chips) == 67)
        assert event.data[0].to_dict()['chips'] == {'50': 1, '10': 1, '1': 7}
        event = await hub.wait_for(EventType.PAYOUT, timeout=5, predicate=lambda e: any([payout.player_id == '69' for payout in e.data]))
        assert len(event.data) == 2
        assert event.data[0].to_dict()['player_id'] in ['42', '69'];
        assert event.data[1].to_dict()['player_id'] in ['42', '69'];
    finally:
        await hub.stop()


@pytest.mark.asyncio
async def test_start_safety():
    hub = EventHub(FakeTransport([], hold_open=True))

    try:
        await hub.start()
        task = hub._task
        await hub.start()
        assert hub._task is task
    finally:
        await hub.stop()


@pytest.mark.asyncio
async def test_stream_filtering():
    hub = EventHub(
        FakeTransport(
            [
                make_event(EventType.WELCOME),
                make_event(EventType.GAME_STATE_UPDATE, {}),
                make_event(EventType.PLAYER_ACTION),
                make_event(EventType.PAYOUT, [{"chips": {'10': 20}, 'player_id': '45'}]),
                make_event(EventType.GAME_STATE_UPDATE, {}),
            ]
        )
    )

    received = [e async for e in hub.stream(EventType.PAYOUT)]

    assert [e.data[0].to_dict()["player_id"] for e in received] == ['45']


@pytest.mark.asyncio
async def test_stream_queued_events():
    # if these were same pot, they would all be in the same single event
    # this test is just testing the queue
    hub = EventHub(FakeTransport([make_event(EventType.PAYOUT, [{"chips": {'10': 5}, 'player_id': str(40 + n)}]) for n in range(5)]))

    received = []
    async for e in hub.stream():
        received.append(e)
        await asyncio.sleep(0.05)  # delay consumption from feed to verify it's still there

    assert [e.data[0].to_dict()["player_id"] for e in received] == ['40', '41', '42', '43', '44']


@pytest.mark.asyncio
async def test_on_event_decorator():
    received: list[Event] = []
    hub = EventHub(
        FakeTransport(
            [
                make_event(EventType.WELCOME),
                make_event(EventType.PAYOUT),
                make_event(EventType.GAME_ENDING),
            ]
        )
    )

    @hub.on_event(EventType.WELCOME, EventType.GAME_ENDING)
    async def on_update(e: Event) -> None:
        received.append(e)

    async with hub:
        await hub.wait_closed(timeout=5)

    assert len(received) == 2


@pytest.mark.asyncio
async def test_on_event_cursed_usage():
    received: list[Event] = []
    hub = EventHub(FakeTransport([make_event(EventType.PAYOUT)]))

    hub.on_event(EventType.PAYOUT)(received.append)

    async with hub:
        await hub.wait_closed(timeout=5)

    assert len(received) == 1


@pytest.mark.asyncio
async def test_subscription_interface_usage():
    game_over_called = asyncio.Event()

    async def on_game_over(e: Event) -> None:
        game_over_called.set()

    hub = EventHub(
        FakeTransport(
            [make_event(EventType.WELCOME), make_event(EventType.GAME_ENDING)],
            hold_open=True,
        )
    )

    try:
        with hub.subscribe(EventType.GAME_ENDING, on_game_over) as subscriber:
            await hub.start()
            await asyncio.wait_for(game_over_called.wait(), timeout=5)

        assert not subscriber.active
        assert hub.subscriber_count(EventType.GAME_ENDING) == 0

        assert subscriber.unsubscribe() is False
    finally:
        await hub.stop()
