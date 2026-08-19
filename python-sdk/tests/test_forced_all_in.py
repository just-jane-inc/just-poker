import random

import helpers as test_helpers
import pytest

import openapi_client as api
import poker_bot.poker_helpers as help
from openapi_client.models.game_table_dto import GameTableDTO
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


@pytest.mark.asyncio
async def test_ante_success():
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
            "10": 38,
            "100": 12,
            "50": 23,
            "500": 2
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "inactive"
        },
        {
        "user_id": "48",
        "display_name": "red",
        "user_type": "bot",
        "position": 1,
        "hole": [],
        "stack": {
            "10": 0,
            "100": 4,
            "50": 0,
            "500": 0
        },
        "current_bet": {},
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
            "10": 2,
            "100": 0,
            "50": 0,
            "500": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "active"
        },
        {
        "user_id": "50",
        "display_name": "rae",
        "user_type": "bot",
        "position": 3,
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
        }
    ],
    "pot": {},
    "street": [],
    "current_round": {
        "bet": 50,
        "current_player_position": 2,
        "current_aggressor": 1,
        "current_round_type": "setup"
    },
    "current_hand": {
        "id": 22, "big_blind": 100, "small_blind": 50, "started_at": "0001-01-01T00:00:00Z"
    },
    "button_position": 1,
    "small_blind_position": 2,
    "big_blind_position": 0
}
"""

    red = test_helpers.get_test_user("red")
    wolf = test_helpers.get_test_user("wolf")
    jane = test_helpers.get_test_user("jane")
    rae = test_helpers.get_test_user("rae")

    game_id = await help.create_game(base_url=test_helpers.base_url, token=jane.token, player_count=4)
    assert game_id

    thing = api.GameApi(help.create_connection(test_helpers.base_url, jane.token))

    red_bot = bot.PokerBot(test_helpers.base_url, red.token, red.user_id, game_id)
    wolf_bot = bot.PokerBot(test_helpers.base_url, wolf.token, wolf.user_id, game_id)
    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)

    await red_bot.join_game()
    await wolf_bot.join_game()
    await jane_bot.join_game()

    await thing.game_game_id_state_post(game_id, GameTableDTO.from_json(json_str=json_str))

    await wolf_bot.ante()
    assert wolf_bot._player.state == "all_in"
