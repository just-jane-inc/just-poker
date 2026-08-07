import pytest
from helpers import base_url, get_test_users

import openapi_client as api
import poker_bot.bot.poker_helpers as help


@pytest.mark.asyncio
async def test_create_game_player_count_too_high_error():
    users = get_test_users()
    try:
        _ = await help.create_game(base_url, str(users[0].token), player_count=10)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


@pytest.mark.asyncio
async def test_create_game_small_blind_negative_error():
    users = get_test_users()
    try:
        _ = await help.create_game(
            base_url, str(users[0].token), player_count=2, bb=50, sb=-100
        )
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


@pytest.mark.asyncio
async def test_create_game_small_blind_too_large_error():
    users = get_test_users()
    try:
        _ = await help.create_game(
            base_url, str(users[0].token), player_count=2, bb=50, sb=100
        )
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024


@pytest.mark.asyncio
async def test_create_game_player_count_too_small_error():
    users = get_test_users()
    try:
        _ = await help.create_game(base_url, str(users[0].token), player_count=-5)
        assert False
    except api.ApiException as ex:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2024
