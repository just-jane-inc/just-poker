import pytest
from helpers import base_url, get_deck, get_test_user, get_test_users

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.api.game_api import GameApi
from openapi_client.models.game_player_intent import GamePlayerIntent
from poker_bot.bot import bot


async def assert_action_throws_with_error_code(action, code: api.JustErrorCode, *args):
    try:
        await action(*args)
        assert False
    except api.ApiException as e:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == code


@pytest.mark.asyncio()
@pytest.mark.timeout(30)
async def test_action_errors():
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
    game_api = GameApi(api_client)
    _ = await game_api.game_game_id_hand_post(
        game_id,
        api.GameGameIdHandPostRequest(api.GameNewHandDTO(deck=get_deck(42))),
    )

    # it is currently bot [2] turn, the bot 3 action here is a violation
    # of the turn order - they must wait for their turn
    await assert_action_throws_with_error_code(
        bots[3].send_action,
        api.JustErrorCode.TurnOrderViolation,
        api.GamePlayerIntent.PlayerIntentAnte,
        {"100": 1},
    )

    # bot 2 must ante the small blind which is currently at 50,
    # a bet of 100 is too much
    await assert_action_throws_with_error_code(
        bots[2].send_action,
        api.JustErrorCode.InvalidBetAmount,
        api.GamePlayerIntent.PlayerIntentAnte,
        {"100": 1},
    )

    # bot 2 must ante right now, a call (event for the correct amount)
    # is not the correct intent
    await assert_action_throws_with_error_code(
        bots[2].send_action,
        api.JustErrorCode.InvalidActionType,
        api.GamePlayerIntent.PlayerIntentCall,
        {"50": 1},
    )

    # you cannot fold when an ante is required
    await assert_action_throws_with_error_code(
        bots[2].send_action,
        api.JustErrorCode.InvalidActionType,
        api.GamePlayerIntent.PlayerIntentFold,
        {},
    )

    await bots[2].ante()
    await bots[3].ante()

    # the intent to ante is only available in the ante round - this intent
    # is now invalid
    await assert_action_throws_with_error_code(
        bots[0].send_action,
        api.JustErrorCode.InvalidActionType,
        api.GamePlayerIntent.PlayerIntentAnte,
        {"100": 1},
    )

    # this player does not have a chip with the denomination of 1000 - it is
    # not valid to bet it. you must make exchanges if you want to bet this
    # specific chip or supply the chips which sum to 1000 if that is your intent
    await assert_action_throws_with_error_code(
        bots[0].send_action,
        api.JustErrorCode.NotEnoughChips,
        api.GamePlayerIntent.PlayerIntentRaise,
        {"1000": 1},
    )

    # a negative chip will always cause an error
    await assert_action_throws_with_error_code(
        bots[0].send_action,
        api.JustErrorCode.InvalidBetAmount,
        api.GamePlayerIntent.PlayerIntentCall,
        {"500": -1},
    )

    await bots[0].fold()
    await bots[1].call()
    await bots[2].call()

    # the big blind has already bet the 100 chips required - they can either
    # check, raise, or fold a 'call' here is invalid
    await assert_action_throws_with_error_code(
        bots[3].call,
        api.JustErrorCode.InvalidActionType,
    )

    await bots[3].check()

    # once folded bot 0 is out of the turn order - they will not be able to move again
    await assert_action_throws_with_error_code(
        bots[0].send_action,
        api.JustErrorCode.TurnOrderViolation,
        api.GamePlayerIntent.PlayerIntentFold,
        {},
    )

    await bots[2].send_action(GamePlayerIntent.PlayerIntentAllIn, {})
    await bots[3].send_action(GamePlayerIntent.PlayerIntentAllIn, {})
    await bots[1].send_action(GamePlayerIntent.PlayerIntentAllIn, {})

    _ = await game_api.game_game_id_hand_post(
        game_id,
        api.GameGameIdHandPostRequest(api.GameNewHandDTO(deck=[])),
    )

    game_state = await bots[0].get_game_state()

    assert game_state.table.players[1].state == "out"
    assert game_state.table.players[3].state == "out"

    await bots[2].send_action(api.GamePlayerIntent.PlayerIntentAnte, {"50": 1})
    await bots[0].send_action(api.GamePlayerIntent.PlayerIntentAnte, {"100": 1})
    await bots[2].send_action(api.GamePlayerIntent.PlayerIntentCall, {"50": 1})

    # await bots[2].ante()
    # await bots[0].ante()
    # await bots[2].call()
