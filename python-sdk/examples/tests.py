import asyncio
import json
import os

import pytest

import openapi_client as api
import src.bot as bot
import src.poker_helpers as help
from openapi_client.models.game_round_type import GameRoundType
from dotenv import load_dotenv

load_dotenv("examples/.env")

base_url = "http://localhost:7653"


@pytest.mark.asyncio
async def test_all_check_2():
    tokens: list[str | None] = [os.getenv("jane"), os.getenv("rae"), os.getenv("red"), os.getenv("wolf")]
    game_id = await help.create_game(base_url, str(tokens[0]))

    bots: list[bot.PokerBot] = []
    for token in tokens:
        b = bot.PokerBot(base_url, token, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()

    # setup
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    # pre flop
    await bots[0].send_action("call", {"10": 10})
    await bots[1].send_action("call", {"10": 10})
    await bots[2].send_action("call", {"10": 5})
    await bots[3].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_flop

    # flop
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})


# TODO: make some test users to re-use rather than making them each time
@pytest.mark.asyncio
async def test_all_check():
    names = ["jane", "wolf", "red", "goblinz"]
    users: list[api.UserApiKey] = []
    for name in names:
        user = await help.create_user(base_url, f"all-check-id-{name}", name)
        assert user
        users.append(user)

    game_id = await help.create_game(base_url, str(users[0].token))

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()

    # setup
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    # pre flop
    await bots[0].send_action("call", {"10": 10})
    await bots[1].send_action("call", {"10": 10})
    await bots[2].send_action("call", {"10": 5})
    await bots[3].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_flop

    # flop
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_turn

    # turn
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_river

    # river
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})


@pytest.mark.asyncio
async def test_all_in():
    names = ["jane", "wolf", "red", "goblinz"]
    users: list[api.UserApiKey] = []
    for name in names:
        user = await help.create_user(base_url, f"all-check-id-{name}", name)
        assert user
        users.append(user)

    game_id = await help.create_game(base_url, str(users[0].token))

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()

    # setup
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    # pre flop
    await bots[0].send_action("all_in", {})
    await bots[1].send_action("all_in", {})
    await bots[2].send_action("all_in", {})
    await bots[3].send_action("all_in", {})

    state = await bots[0].get_game_state()

    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_completed
