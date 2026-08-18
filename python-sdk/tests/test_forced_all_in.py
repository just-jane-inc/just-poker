import random

import helpers as test_helpers
import pytest

# from helpers import base_url, get_test_users, make_tests_work_for_fricking_windows
import openapi_client as api
import poker_bot.bot.poker_helpers as help
from poker_bot.bot import bot

test_helpers.make_tests_work_for_fricking_windows()

json_str = """
{
    "players": [
      {
        "user_id": "45",
        "display_name": "jane",
        "user_type": "bot",
        "position": 0,
        "hole": [],
        "stack": {
          "10": 0,
          "100": 0,
          "50": 0,
          "500": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
      },
      {
        "user_id": "48",
        "display_name": "red",
        "user_type": "bot",
        "position": 1,
        "hole": [],
        "stack": {
          "10": 30,
          "100": 9,
          "50": 17,
          "500": 2
        },
        "current_bet": {
          "50": 1
        },
        "pot_contribution": 0,
        "state": "inactive"
      },
      {
        "user_id": "49",
        "display_name": "wolf",
        "user_type": "bot",
        "position": 2,
        "hole": [],
        "stack": {
          "10": 0,
          "100": 0,
          "50": 1,
          "500": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "active"
      }
    ],
    "pot": {
      "50": 1
    },
    "street": [],
    "current_round": {
      "bet": 100,
      "current_player_position": 2,
      "current_aggressor": 1,
      "current_round_type": "setup"
    },
    "current_hand": {
      "id": 13,
      "big_blind": 100,
      "small_blind": 50,
      "started_at": "0001-01-01T00:00:00Z"
    },
    "button_position": 1,
    "small_blind_position": 1,
    "big_blind_position": 2
}
"""


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

    if random.rand() > 0.5:
        await wolf_bot.ante()
    else:
        await wolf_bot.all_in()
