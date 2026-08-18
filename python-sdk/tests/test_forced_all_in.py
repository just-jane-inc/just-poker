import random

import helpers as test_helpers
import pytest

import openapi_client as api
import poker_bot.poker_helpers as help
from poker_bot import bot

test_helpers.make_tests_work_for_fricking_windows()


@pytest.mark.asyncio
async def test_ante_error():
    red = test_helpers.get_test_user("red")
    wolf = test_helpers.get_test_user("wolf")

    game_id = await help.create_game(
        base_url=test_helpers.base_url, token=str(red.token), player_count=2, chips={"100": 5, "50": 1, "10": 5}
    )
    assert game_id

    api.GameApi()

    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)
    wolf_bot = bot.PokerBot(test_helpers.base_url, wolf.token, wolf.user_id, game_id)

    await red_bot.join_game()
    await wolf_bot.join_game()
    await red_bot.start_game()

    await wolf_bot.ante()
    await red_bot.ante()
    await wolf_bot.raise_bet(550)
    await red_bot.call()
    await red_bot.check()
    await wolf_bot.fold()

    await red_bot.ante()

    try:
        await wolf_bot.send_action("ante", {"100": 1})
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == api.JustErrorCode.NotEnoughChips

    try:
        await wolf_bot.send_action("ante", {"10": 3})
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == api.JustErrorCode.InvalidBetAmount

    try:
        await wolf_bot.send_action("ante", {"10": 6})
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == api.JustErrorCode.InvalidBetAmount

    if random.random() > 0.5:
        await wolf_bot.ante()
    else:
        await wolf_bot.all_in()
