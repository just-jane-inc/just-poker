import json

import pytest
from helpers import base_url, get_deck, get_test_user, get_test_users

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.models.game_game_id_hand_post_request import (
    GameGameIdHandPostRequest,
)
from openapi_client.models.game_new_hand_dto import GameNewHandDTO
from poker_bot.bot import bot


@pytest.mark.asyncio
async def test_post_new_hand_with_deck():
    jill = get_test_user("jill")

    api_client = help.create_connection(base_url, jill.token)

    game_id = await help.create_game(base_url, jill.token, auto_start_hands=False)
    assert game_id

    users = get_test_users()
    assert len(users) == 4
    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await help.start_game(api_client, game_id)
    game_api = api.GameApi(api_client)
    resp = await game_api.game_game_id_hand_post(
        game_id, GameGameIdHandPostRequest(GameNewHandDTO(deck=get_deck(42)))
    )

    print(json.dumps(resp, indent=2))

    await bots[2].ante()
    await bots[3].ante()

    await bots[0].call()
    await bots[1].call()
    await bots[2].call()
    await bots[3].check()

    jill_bot = bot.PokerBot(base_url, jill.token, jill.user_id, game_id)
    game_state = await jill_bot.get_game_state()
    assert game_state

    assert game_state.table.street[0].rank == 54
    assert game_state.table.street[0].suit == 100

    assert game_state.table.street[1].rank == 52
    assert game_state.table.street[1].suit == 100

    assert game_state.table.street[2].rank == 75
    assert game_state.table.street[2].suit == 99


@pytest.mark.asyncio
async def test_post_new_hand_with_deck_error():
    jill = get_test_user("jill")

    api_client = help.create_connection(base_url, jill.token)

    game_id = await help.create_game(base_url, jill.token, auto_start_hands=False)
    assert game_id

    users = get_test_users()
    assert len(users) == 4
    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await help.start_game(api_client, game_id)
    game_api = api.GameApi(api_client)
    resp = await game_api.game_game_id_hand_post(
        game_id, GameGameIdHandPostRequest(GameNewHandDTO(deck=get_deck(42)))
    )

    print(json.dumps(resp, indent=2))

    await bots[2].ante()
    await bots[3].ante()

    try:
        _ = await game_api.game_game_id_hand_post(
            game_id, GameGameIdHandPostRequest(GameNewHandDTO(deck=get_deck(42)))
        )
    except api.ApiException as e:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2025
