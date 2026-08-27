import json

import helpers as test_helpers
import pytest

import openapi_client as api
import poker_bot.poker_helpers as help
from poker_bot import bot

test_helpers.make_tests_work_for_fricking_windows()


test_1 = """
{
    "players": [
        {
        "user_id": "48",
        "display_name": "red",
        "user_type": "bot",
        "position": 0,
        "hole": [
            {
            "rank": 65,
            "suit": 100
            },
            {
            "rank": 81,
            "suit": 104
            }
        ],
        "stack": {
            "10": 0,
            "50": 0,
            "100": 0,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 660,
        "state": "all_in"
        },
        {
        "user_id": "68",
        "display_name": "genet-test1",
        "user_type": "bot",
        "position": 1,
        "hole": [],
        "stack": {
            "10": 0,
            "50": 0,
            "100": 0,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
        },
        {
        "user_id": "69",
        "display_name": "genet-test2",
        "user_type": "bot",
        "position": 2,
        "hole": [],
        "stack": {
            "10": 0,
            "50": 0,
            "100": 0,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
        },
        {
        "user_id": "49",
        "display_name": "wolf",
        "user_type": "bot",
        "position": 3,
        "hole": [
            {
            "rank": 74,
            "suit": 115
            },
            {
            "rank": 65,
            "suit": 99
            }
        ],
        "stack": {
            "10": 101,
            "50": 36,
            "100": 2,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 950,
        "state": "active"
        },
        {
        "user_id": "45",
        "display_name": "jane",
        "user_type": "bot",
        "position": 4,
        "hole": [
            {
            "rank": 52,
            "suit": 104
            },
            {
            "rank": 55,
            "suit": 104
            }
        ],
        "stack": {
            "10": 58,
            "50": 13,
            "100": 37,
            "500": 3,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 950,
        "state": "inactive"
        },
        {
        "user_id": "72",
        "display_name": "genet-test5",
        "user_type": "bot",
        "position": 5,
        "hole": [],
        "stack": {
            "10": 0,
            "50": 0,
            "100": 0,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
        },
        {
        "user_id": "73",
        "display_name": "genet-test6",
        "user_type": "bot",
        "position": 6,
        "hole": [],
        "stack": {
            "10": 0,
            "50": 0,
            "100": 0,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
        },
        {
        "user_id": "74",
        "display_name": "genet-test7",
        "user_type": "bot",
        "position": 7,
        "hole": [],
        "stack": {
            "10": 0,
            "50": 0,
            "100": 0,
            "500": 0,
            "1000": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
        }
    ],
    "pot": {
        "10": 1,
        "50": 3,
        "100": 19,
        "500": 1,
        "1000": 0
    },
    "street": [
        {
        "rank": 53,
        "suit": 99
        },
        {
        "rank": 51,
        "suit": 100
        },
        {
        "rank": 51,
        "suit": 99
        },
        {
        "rank": 75,
        "suit": 100
        },
        {
        "rank": 57,
        "suit": 104
        }
    ],
    "current_round": {
        "bet": 0,
        "current_player_position": 3,
        "current_aggressor": 4,
        "current_round_type": "river"
    },
    "current_hand": {
        "id": 17,
        "big_blind": 100,
        "small_blind": 50,
        "started_at": "0001-01-01T00:00:00Z"
    },
    "button_position": 0,
    "small_blind_position": 3,
    "big_blind_position": 4
}
"""


@pytest.mark.asyncio
async def test_load_game_from_state():
    jill = test_helpers.get_test_user("jill")

    api_client = help.create_connection(test_helpers.base_url, jill.token)

    game_id = await help.create_game(
        test_helpers.base_url,
        jill.token,
        auto_start_hands=True,
        chips={"10": 20, "50": 6, "100": 5, "500": 1, "1000": 0},
        denominations=[10, 50, 100, 500, 1000],
        player_count=8,
    )

    assert game_id

    game_api: api.GameApi = api.GameApi(api_client)

    data = json.loads(test_1)
    table_dto: api.GameTableDTO = api.GameTableDTO.from_dict(data)
    assert table_dto is not None
    await game_api.game_game_id_state_post(game_id, table_dto)

    wolf = test_helpers.get_test_user("wolf")
    wolf_bot = bot.PokerBot(test_helpers.base_url, wolf.token, wolf.user_id, game_id)
    await wolf_bot.get_game_state()
    await wolf_bot.check()
    await wolf_bot.get_game_state()
    assert wolf_bot.chip_total() == 3590


test_2 = """
{
  "game_config": {
    "player_count": 8,
    "starting_chips": {
      "10": 20,
      "100": 5,
      "1000": 0,
      "50": 6,
      "500": 1
    },
    "big_blind": 100,
    "small_blind": 50,
    "auto_starts_hands": true,
    "chip_denominations": [
      10,
      50,
      100,
      500,
      1000
    ],
    "bot_turn_timeout": 0
  },
  "table": {
    "players": [
      {
        "user_id": "67",
        "display_name": "genet-test",
        "user_type": "bot",
        "position": 0,
        "hole": [
          {
            "rank": 53,
            "suit": 99
          },
          {
            "rank": 81,
            "suit": 115
          }
        ],
        "stack": {
          "10": 10,
          "100": 2,
          "1000": 0,
          "50": 6,
          "500": 1
        },
        "current_bet": {},
        "pot_contribution": 50,
        "state": "folded"
      },
      {
        "user_id": "68",
        "display_name": "genet-test1",
        "user_type": "bot",
        "position": 1,
        "hole": [],
        "stack": {
          "10": 0,
          "100": 0,
          "1000": 0,
          "50": 0,
          "500": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
      },
      {
        "user_id": "69",
        "display_name": "genet-test2",
        "user_type": "bot",
        "position": 2,
        "hole": [
          {
            "rank": 65,
            "suit": 99
          },
          {
            "rank": 52,
            "suit": 99
          }
        ],
        "stack": {
          "10": 3,
          "100": 0,
          "1000": 0,
          "50": 4,
          "500": 0
        },
        "current_bet": {
          "10": 4,
          "100": 4
        },
        "pot_contribution": 280,
        "state": "inactive"
      },
      {
        "user_id": "70",
        "display_name": "genet-test3",
        "user_type": "bot",
        "position": 3,
        "hole": [],
        "stack": {
          "10": 0,
          "100": 0,
          "1000": 0,
          "50": 0,
          "500": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
      },
      {
        "user_id": "71",
        "display_name": "genet-test4",
        "user_type": "bot",
        "position": 4,
        "hole": [
          {
            "rank": 52,
            "suit": 100
          },
          {
            "rank": 65,
            "suit": 104
          }
        ],
        "stack": {
          "10": 48,
          "100": 14,
          "1000": 2,
          "50": 14,
          "500": 1
        },
        "current_bet": {
          "10": 4,
          "100": 4
        },
        "pot_contribution": 280,
        "state": "inactive"
      },
      {
        "user_id": "72",
        "display_name": "genet-test5",
        "user_type": "bot",
        "position": 5,
        "hole": [],
        "stack": {
          "10": 0,
          "100": 0,
          "1000": 0,
          "50": 0,
          "500": 0
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "out"
      },
      {
        "user_id": "45",
        "display_name": "jane",
        "user_type": "bot",
        "position": 6,
        "hole": [
          {
            "rank": 53,
            "suit": 104
          },
          {
            "rank": 84,
            "suit": 104
          }
        ],
        "stack": {
          "10": 17,
          "100": 0,
          "1000": 0,
          "50": 4,
          "500": 1
        },
        "current_bet": {},
        "pot_contribution": 280,
        "state": "active"
      },
      {
        "user_id": "74",
        "display_name": "genet-test7",
        "user_type": "bot",
        "position": 7,
        "hole": [
          {
            "rank": 51,
            "suit": 100
          },
          {
            "rank": 74,
            "suit": 99
          }
        ],
        "stack": {
          "10": 20,
          "100": 10,
          "1000": 0,
          "50": 5,
          "500": 3
        },
        "current_bet": {},
        "pot_contribution": 0,
        "state": "folded"
      }
    ],
    "pot": {
      "10": 17,
      "100": 14,
      "50": 4
    },
    "street": [
      {
        "rank": 56,
        "suit": 104
      },
      {
        "rank": 51,
        "suit": 99
      },
      {
        "rank": 53,
        "suit": 100
      },
      {
        "rank": 50,
        "suit": 104
      },
      {
        "rank": 53,
        "suit": 115
      }
    ],
    "current_round": {
      "bet": 440,
      "current_player_position": 6,
      "current_aggressor": 2,
      "current_round_type": "river"
    },
    "current_hand": {
      "id": 5,
      "big_blind": 100,
      "small_blind": 50,
      "started_at": "0001-01-01T00:00:00Z"
    },
    "button_position": 7,
    "small_blind_position": 0,
    "big_blind_position": 2
  }
}
"""


@pytest.mark.asyncio
async def test_game_with_remainder():
    def sum_stack(stack: dict[str, int]):
        return sum(int(d) * c for d, c in stack.items())

    jill = test_helpers.get_test_user("jill")
    api_client = help.create_connection(test_helpers.base_url, jill.token)

    data = json.loads(test_2)
    game_dto: api.GameGameDTO = api.GameGameDTO.from_dict(data)
    game_id = await help.create_game(
        test_helpers.base_url,
        jill.token,
        auto_start_hands=True,
        chips=game_dto.game_config.starting_chips,
        denominations=game_dto.game_config.chip_denominations,
        player_count=game_dto.game_config.player_count,
    )

    assert game_id
    game_api: api.GameApi = api.GameApi(api_client)
    await game_api.game_game_id_state_post(game_id, game_dto.table)

    jane = test_helpers.get_test_user("jane")
    jane_bot = bot.PokerBot(test_helpers.base_url, jane.token, jane.user_id, game_id)
    game_state = await jane_bot.get_game_state()
    position_two = sum_stack(game_state.table.players[2].stack)
    position_four = sum_stack(game_state.table.players[4].stack)
    await jane_bot.call()

    game_state = await jane_bot.get_game_state()
    assert position_two + 1110 == sum_stack(game_state.table.players[2].stack)
    assert position_four + 1100 == sum_stack(game_state.table.players[4].stack)
