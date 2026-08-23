import random
import asyncio

import helpers as test_helpers
import pytest

import poker_bot.poker_helpers as help
from openapi_client import GameChipExchangeDTO
from poker_bot import Event, EventType, bot

test_helpers.make_tests_work_for_fricking_windows()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_chip_exchange_happens():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    starting_chips = {
        "10": 1,
        "50": 0,
        "100": 0,
        "500": 1,
        "1000": 1,
    }

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=False,
                                     chips=starting_chips, denominations=[10, 50, 100, 500, 1000])
    assert game_id
    print(game_id)

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    received: list[GameChipExchangeDTO] = []

    @jane_bot.events.on_event(EventType.CHIP_EXCHANGE)
    async def on_chip_exchange(e: Event):
        received.append(e.data)

    await jane_bot.start_events()
    await jane_bot.get_game_state()

    for x in range(5):
        target = random.randint(2, 98) * 10 # 20 to 98
        await jane_bot._compute_valid_bet(target)
        await asyncio.sleep(0.05)

    assert len(received) > 0

@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_raise_with_big_chips():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    starting_chips = {
        "10": 20,
        "50": 6,
        "100": 5,
        "500": 1,
        "1000": 0,
    }

    total = sum(int(d) * c for d, c in starting_chips.items())

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=True,
                                     chips=starting_chips, denominations=[10, 50, 100, 500, 1000])
    assert game_id
    print(game_id)

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    await jane_bot.start_events()
    await jane_bot.start_game()

    await red_bot.ante()
    await jane_bot.ante()
    await red_bot.call()
    assert jane_bot.chip_total() > 1100

    error_raised = False
    try:
        await jane_bot.raise_bet(1100)
    except:
        error_raised = True
    assert not error_raised


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_bet_with_big_chips():
    jane = test_helpers.get_test_user("jane")
    red = test_helpers.get_test_user("red")

    starting_chips = {
        "10": 20,
        "50": 6,
        "100": 5,
        "500": 1,
        "1000": 0,
    }

    total = sum(int(d) * c for d, c in starting_chips.items())

    game_id = await help.create_game(test_helpers.base_url, str(jane.token), auto_start_hands=True,
                                     chips=starting_chips, denominations=[10, 50, 100, 500, 1000])
    assert game_id
    print(game_id)

    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)

    await jane_bot.join_game()
    await red_bot.join_game()

    await jane_bot.start_events()
    await jane_bot.get_game_state()

    for target in (100, 500, 900, 1000, 1100, 1200, 1400, total):
        bet = await jane_bot._compute_valid_bet(target)

        assert sum(int(d) * c for d, c in bet.items()) == target, f"bet {bet} is not worth {target}"

        held = jane_bot._player.stack
        short = {d: c - held.get(d, 0) for d, c in bet.items() if c > held.get(d, 0)}
        assert not short, f"betting {target} returned {bet}; needs {short} more than the player has ({held})"
