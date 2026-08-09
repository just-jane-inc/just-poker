import pytest
from helpers import base_url, get_test_users

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from poker_bot.bot import bot


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
    await bots[2].ante()
    await bots[3].ante()

    # pre flop
    await bots[0].call()
    await bots[1].raise_bet(200)
    await bots[2].fold()
    await bots[3].call()
    await bots[0].call()

    state = await bots[0].get_game_state()
    assert (
        state.table.current_round.current_round_type == api.GameRoundType.RoundTypeFlop
    )

    # flop
    await bots[3].check()
    await bots[0].check()
    await bots[1].check()

    # turn
    await bots[3].check()
    await bots[0].check()
    await bots[1].check()

    # river
    await bots[3].check()
    await bots[0].check()
    await bots[1].check()

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
    await bots[2].ante()
    await bots[3].ante()

    # pre flop
    await bots[0].all_in()
    await bots[1].all_in()
    await bots[2].all_in()
    await bots[3].all_in()


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
    state = await bots[0].get_game_state()
    assert state.table.current_hand.id == 1

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
        state.table.current_round.current_round_type == api.GameRoundType.RoundTypeTurn
    )

    # turn
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert (
        state.table.current_round.current_round_type == api.GameRoundType.RoundTypeRiver
    )

    # river
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_hand.id == 2
