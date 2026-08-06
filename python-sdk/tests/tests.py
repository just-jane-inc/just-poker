import os

import pytest
from dotenv import load_dotenv

import openapi_client as api
import poker_bot.bot.poker_helpers as help
import poker_bot.tools.tui.setup_tui_example as examples
from poker_bot.bot import bot

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")


def get_test_users() -> list[examples.TestUser]:
    test_users = examples.get_test_users()
    users = []
    for user in test_users:
        if user.username == "jill":
            continue
        users.append(user)

    return users


@pytest.mark.asyncio
async def test_turn_order_violation():
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

    # setup
    try:
        _ = await bots[0].send_action("ante", {"10": 5})
        assert False
    except api.exceptions.ApiException as ex:
        body = api.JustResponseMessageJustErrorDTO.from_json(ex.body)
        assert body.type == "error"
        assert body.data.error_code == 2000


@pytest.mark.asyncio
async def test_game_one():
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

    # setup
    await bots[2].send_action("ante", {"50": 1})
    await bots[3].send_action("ante", {"10": 10})

    # pre flop
    await bots[0].send_action("call", {"50": 2})
    await bots[1].send_action("raise", {"100": 2})
    await bots[2].send_action("fold", {})
    await bots[3].send_action("call", {"100": 1})
    await bots[0].send_action("call", {"100": 1})

    state = await bots[0].get_game_state()
    assert (
        state.table.current_round.current_round_type == api.GameRoundType.RoundTypeFlop
    )

    # flop
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    # turn
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    # river
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state.table.current_hand.id == 2


@pytest.mark.asyncio
async def test_all_in_game():
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

    # setup
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    # pre flop
    await bots[0].send_action("all_in", {})
    await bots[1].send_action("all_in", {})
    await bots[2].send_action("all_in", {})
    await bots[3].send_action("all_in", {})

    try:
        _ = await bots[0].get_game_state()
        assert False
    except api.ApiException:
        assert True


@pytest.mark.asyncio
async def test_checking_game_one():
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
    state = bots[0].get_game_state()
    assert state.table.current_hand.id == 0

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
    assert (
        state.table.current_round.current_round_type == api.GameRoundType.RoundTypeFlop
    )

    # flop
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert (
        state.table.current_round.current_round_type == api.GameRoundType.RoundTypeturn
    )

    # turn
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert (
        state.table.current_round.current_round_type
        == api.api.GameRoundType.RoundTypeRiver
    )

    # river
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    assert state.table.current_hand.id == 1
