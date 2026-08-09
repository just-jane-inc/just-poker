import asyncio

import pytest
from helpers import base_url, make_tests_work_for_fricking_windows, get_test_user

import poker_bot.bot.poker_helpers as help
from poker_bot.bot import bot, EventType, Event

make_tests_work_for_fricking_windows()

@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over():
    jane = get_test_user("jane")
    red = get_test_user("red")

    game_id = await help.create_game(base_url, str(jane.token))
    assert game_id
    print(game_id)

    jane_bot = bot.PokerBot(base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    received: list[Event] = []

    @jane_bot.on_event(EventType.GAME_OVER)
    async def on_game_over(e: Event):
        received.append(e)

    await jane_bot.start_events() # If you subscribe, you are responsible for starting listening
    # But red_bot doesn't need to, because this is implicitly started in the helper functions for send action
    # as they wait for turn, which needs this input
    # AKA The above is not needed if you don't use send_action directly like this
    # which is why red_bot is chillin'

    await jane_bot.start_game()
    await red_bot.ante()
    await jane_bot.send_action("ante", {"100": 1})
    await red_bot.all_in()
    await jane_bot.send_action("all_in", {})
    await asyncio.sleep(3)


    assert len(received) == 1
    e = received[0]
    assert e is not None
    assert e.event_type == EventType.GAME_OVER
    assert e.data is not None and isinstance(e.data, list) and len(e.data) > 0

    await jane_bot.stop_events()
    await red_bot.stop_events()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over_via_subscribe():
    jane = get_test_user("jane")
    red = get_test_user("red")

    game_id = await help.create_game(base_url, str(jane.token))
    assert game_id

    jane_bot = bot.PokerBot(base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    received: list[Event] = []

    async def on_game_over(e: Event):
        received.append(e)

    jane_bot.subscribe(EventType.GAME_OVER, on_game_over)
    await jane_bot.start_events() # If you subscribe, you are responsible for starting listening

    await jane_bot.start_game()

    await red_bot.ante()
    await jane_bot.send_action("ante", {"100": 1})

    await red_bot.all_in()
    await jane_bot.send_action("all_in", {})
    await asyncio.sleep(3)

    assert len(received) == 1
    e = received[0]
    assert e is not None
    assert e.event_type == EventType.GAME_OVER
    assert e.data is not None and isinstance(e.data, list) and len(e.data) > 0

    await jane_bot.stop_events()
    await red_bot.stop_events()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over_better_closure():
    jane = get_test_user("jane")
    red = get_test_user("red")

    game_id = await help.create_game(base_url, str(jane.token))
    assert game_id

    jane_bot = bot.PokerBot(base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(base_url, red.token, red.user_id, game_id)
    await jane_bot.join_game()
    await red_bot.join_game()

    async with jane_bot, red_bot:
        # Avoids need to ensure events need starting, regardless of what is later used

        game_over = asyncio.Event()
        received: list[Event] = []

        @jane_bot.on_event(EventType.GAME_OVER)
        async def on_game_over(e: Event):
            received.append(e)
            game_over.set()

        await jane_bot.start_game()

        await red_bot.send_action("ante", {"50": 1})
        await jane_bot.ante()
        await red_bot.send_action("all_in", {})
        await jane_bot.all_in()

        await asyncio.wait_for(game_over.wait(), timeout=10)

    assert len(received) == 1
    e = received[0]
    assert e.event_type == EventType.GAME_OVER
    assert isinstance(e.data, list) and len(e.data) > 0
