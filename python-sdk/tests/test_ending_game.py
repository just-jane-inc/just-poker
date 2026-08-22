import asyncio

import helpers as test_helpers
import pytest

import openapi_client as api
import poker_bot.poker_helpers as help
from openapi_client.api.game_api import GameApi
from openapi_client.models.game_card_dto import GameCardDTO
from poker_bot import Event, EventType, bot

test_helpers.make_tests_work_for_fricking_windows()


async def start_hand(game_id: str, user):
    deck = test_helpers.fill_deck_remainder(
        [
            GameCardDTO(rank=help.CardRank.ACE.value, suit=help.CardSuit.HEART.value),
            GameCardDTO(rank=help.CardRank.TWO.value, suit=help.CardSuit.HEART.value),
            GameCardDTO(rank=help.CardRank.ACE.value, suit=help.CardSuit.DIAMOND.value),
            GameCardDTO(rank=help.CardRank.TWO.value, suit=help.CardSuit.CLUB.value),
            GameCardDTO(rank=help.CardRank.KING.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.ACE.value, suit=help.CardSuit.CLUB.value),
            GameCardDTO(rank=help.CardRank.ACE.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.NINE.value, suit=help.CardSuit.SPADE.value),
        ]
    )

    game_api = GameApi(help.create_connection(test_helpers.base_url, user.token))
    assert await game_api.game_game_id_hand_post(game_id, api.GameNewHandDTO(deck=deck))


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=False)
    assert game_id
    print(game_id)

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    received: list[Event] = []

    @jane_bot.events.on_event(EventType.GAME_ENDING)
    async def on_game_over(e: Event):
        received.append(e)

    # If you subscribe, you are responsible for starting listening
    await jane_bot.start_events()
    # But red_bot doesn't need to, because this is implicitly started in the helper functions for send action
    # as they wait for turn, which needs this input
    # AKA The above is not needed if you don't use send_action directly like this
    # which is why red_bot is chillin'

    await jane_bot.start_game()
    await start_hand(game_id, jane)

    await red_bot.ante()
    await jane_bot.send_action("ante", {"100": 1})
    await red_bot.all_in()
    await jane_bot.send_action("all_in", {})

    # this is self documenting, the game does not end until a hand begins, duh
    await asyncio.sleep(3)

    assert len(received) == 1
    e = received[0]
    assert e is not None
    assert e.event_type == EventType.GAME_ENDING
    assert e.data is not None and isinstance(e.data, list) and len(e.data) > 0

    await jane_bot.stop_events()
    await red_bot.stop_events()

    # We basically attach an unsubscribe function to each callback fn during attaching
    on_game_over.unsubscribe()

    assert jane_bot._hub is not None
    assert jane_bot._hub.subscriber_count(EventType.GAME_ENDING) == 0


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over_via_subscribe():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=False)
    assert game_id

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    received: list[Event] = []

    async def on_game_over(e: Event):
        received.append(e)

    async def on_player_action(e: Event):
        pass

    subscriber = jane_bot.events.subscribe(EventType.GAME_ENDING, on_game_over)
    subscriber2 = jane_bot.events.subscribe(EventType.PLAYER_ACTION, on_player_action)
    await jane_bot.start_events()

    # If you subscribe, you are responsible for starting listening
    assert jane_bot._hub is not None
    assert jane_bot._hub.subscriber_count(EventType.PLAYER_ACTION) == 1

    with subscriber2:
        ## just to invoke its unsubscribe
        pass
    assert jane_bot._hub.subscriber_count(EventType.PLAYER_ACTION) == 0

    await jane_bot.start_game()
    await start_hand(game_id, jane)

    await red_bot.ante()
    await jane_bot.send_action("ante", {"100": 1})

    await red_bot.all_in()
    await jane_bot.send_action("all_in", {})

    await asyncio.sleep(3)

    assert len(received) == 1
    e = received[0]
    assert e is not None
    assert e.event_type == EventType.GAME_ENDING
    assert e.data is not None and isinstance(e.data, list) and len(e.data) > 0

    assert jane_bot._hub is not None
    assert jane_bot._hub.subscriber_count(EventType.GAME_ENDING) == 1

    subscriber.unsubscribe()
    assert jane_bot._hub is not None
    assert jane_bot._hub.subscriber_count(EventType.GAME_ENDING) == 0

    # show it noops on unsubscribing already unsubbed
    subscriber2.unsubscribe()
    assert jane_bot._hub.subscriber_count(EventType.PLAYER_ACTION) == 0

    await jane_bot.stop_events()
    await red_bot.stop_events()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over_via_subscribe_and_unhook():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=False)
    assert game_id

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    received: list[Event] = []

    async def on_game_over(e: Event):
        received.append(e)

    # Do not need to unsubscribe if using with syntax
    with jane_bot.events.subscribe(EventType.GAME_ENDING, on_game_over):
        # If you subscribe, you are responsible for starting listening
        await jane_bot.start_events()

        await jane_bot.start_game()
        await start_hand(game_id, jane)

        assert jane_bot._hub is not None
        assert jane_bot._hub.subscriber_count(EventType.GAME_ENDING) == 1

        await red_bot.ante()
        await jane_bot.send_action("ante", {"100": 1})

        await red_bot.all_in()
        await jane_bot.send_action("all_in", {})

        await asyncio.sleep(3)

    assert len(received) == 1
    e = received[0]
    assert e is not None
    assert e.event_type == EventType.GAME_ENDING
    assert e.data is not None and isinstance(e.data, list) and len(e.data) > 0

    assert jane_bot._hub is not None
    assert jane_bot._hub.subscriber_count(EventType.GAME_ENDING) == 0

    await jane_bot.stop_events()
    await red_bot.stop_events()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_receive_game_over_better_closure():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=False)
    assert game_id

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)
    await jane_bot.join_game()
    await red_bot.join_game()

    async with jane_bot, red_bot:
        # Avoids need to ensure events need starting, regardless of what is later used
        # As start_events is in the aenter

        game_over = asyncio.Event()
        received: list[Event] = []

        @jane_bot.events.on_event(EventType.GAME_ENDING)
        async def on_game_over(e: Event):
            received.append(e)
            game_over.set()

        await jane_bot.start_game()
        await start_hand(game_id, jane)

        await red_bot.send_action("ante", {"50": 1})
        await jane_bot.ante()
        await red_bot.send_action("all_in", {})
        await jane_bot.all_in()

        await asyncio.wait_for(game_over.wait(), timeout=10)

        # still need to unsubscribe as aexit doesn't track each
        on_game_over.unsubscribe()

    assert len(received) == 1
    e = received[0]
    assert e.event_type == EventType.GAME_ENDING
    assert isinstance(e.data, list) and len(e.data) > 0

    assert jane_bot.events is not None
    assert jane_bot.events.subscriber_count(EventType.GAME_ENDING) == 0
