import helpers as test_helpers
import pytest

import openapi_client as api
import poker_bot.poker_helpers as help
from openapi_client.models.just_error_code import JustErrorCode

test_helpers.make_tests_work_for_fricking_windows()


@pytest.mark.asyncio
async def test_create_game_player_count_too_high_error():
    users = test_helpers.get_test_users()
    try:
        _ = await help.create_game(test_helpers.base_url, str(users[0].token), player_count=10)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


@pytest.mark.asyncio
async def test_create_game_small_blind_negative_error():
    users = test_helpers.get_test_users()
    try:
        _ = await help.create_game(test_helpers.base_url, str(users[0].token), player_count=2, bb=50, sb=-100)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


@pytest.mark.asyncio
async def test_create_game_small_blind_too_large_error():
    users = test_helpers.get_test_users()
    try:
        _ = await help.create_game(test_helpers.base_url, str(users[0].token), player_count=2, bb=50, sb=100)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


@pytest.mark.asyncio
async def test_create_game_player_count_too_small_error():
    users = test_helpers.get_test_users()
    try:
        _ = await help.create_game(test_helpers.base_url, str(users[0].token), player_count=-5)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


invalid_starting_chips_data = [
    ({"-10": 5}, [-10], JustErrorCode.InvalidChipDenomination),
    ({"10": 5}, [5], JustErrorCode.InvalidGameConfiguration),
    ({"5": -1}, [5], JustErrorCode.InvalidChipCount),
]


@pytest.mark.parametrize("chips,denominations,err", invalid_starting_chips_data)
@pytest.mark.asyncio
async def test_create_game_invalid_starting_chips(chips, denominations, err):
    users = test_helpers.get_test_users()
    try:
        _ = await help.create_game(test_helpers.base_url, str(users[0].token), chips=chips, denominations=denominations)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == err
