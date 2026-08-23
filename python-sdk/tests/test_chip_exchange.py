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