import pytest
from helpers import base_url, get_test_users

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from poker_bot.bot import bot


@pytest.mark.asyncio
async def test_raise_does_not_equal_all_in():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    try:
        _ = await bots[0].raise_bet(2000)
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2003


@pytest.mark.asyncio
async def test_raise_amount_too_low():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    try:
        _ = await bots[0].send_action("raise", {"100": 1})
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2003


@pytest.mark.asyncio
async def test_call_amount_too_large():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    try:
        _ = await bots[0].send_action("call", {"50": 3})
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2003


@pytest.mark.asyncio
async def test_call_amount_too_low():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    try:
        _ = await bots[0].send_action("call", {"10": 1})
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2003


@pytest.mark.asyncio
async def test_ante_amount_too_large():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()

    try:
        _ = await bots[2].send_action("ante", {"100": 1})
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2003


@pytest.mark.asyncio
async def test_ante_amount_too_small():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()

    try:
        _ = await bots[2].send_action("ante", {"10": 1})
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2003
