import json

import aiofiles
import helpers as test_helpers
import pytest

import openapi_client as api
import poker_bot.poker_helpers as help

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
async def test_load_game_from_state():
    jill = test_helpers.get_test_user("jill")

    api_client = help.create_connection(test_helpers.base_url, jill.token)

    game_id = await help.create_game(test_helpers.base_url, jill.token, auto_start_hands=False)
    assert game_id

    game_api: api.GameApi = api.GameApi(api_client)

    async with aiofiles.open("tests/data/load_game_test_data_1.json", "r") as f:
        json_str = await f.read()
        data = json.loads(json_str)

    table_dto: api.GameTableDTO = api.GameTableDTO.from_dict(data)
    assert table_dto is not None
    await game_api.game_game_id_state_post(game_id, table_dto)
