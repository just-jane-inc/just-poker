import json

import pytest
from helpers import (
    base_url,
    fill_deck_remainder,
    get_deck,
    get_test_user,
    get_test_users,
    make_tests_work_for_fricking_windows,
)

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.models.game_card_dto import GameCardDTO
from openapi_client.models.game_new_hand_dto import GameNewHandDTO
from poker_bot.bot import bot

make_tests_work_for_fricking_windows()


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
    deck = fill_deck_remainder(
        [
            GameCardDTO(rank=ord("A"), suit=ord("d")),  # rae
            GameCardDTO(rank=ord("A"), suit=ord("h")),  # jane
            GameCardDTO(rank=ord("2"), suit=ord("h")),  # red
            GameCardDTO(rank=ord("2"), suit=ord("d")),  # wolf
            GameCardDTO(rank=ord("K"), suit=ord("h")),  # rae
            GameCardDTO(rank=ord("K"), suit=ord("d")),  # jane
            GameCardDTO(rank=ord("7"), suit=ord("h")),  # red
            GameCardDTO(rank=ord("7"), suit=ord("d")),  # wolf
            GameCardDTO(rank=ord("3"), suit=ord("s")),  # burn
            GameCardDTO(rank=ord("A"), suit=ord("s")),  # flop
            GameCardDTO(rank=ord("A"), suit=ord("c")),  # flop
            GameCardDTO(rank=ord("K"), suit=ord("c")),  # flop
            GameCardDTO(rank=ord("3"), suit=ord("d")),  # burn
            GameCardDTO(rank=ord("5"), suit=ord("h")),  # turn
            GameCardDTO(rank=ord("3"), suit=ord("c")),  # burn
            GameCardDTO(rank=ord("6"), suit=ord("h")),  # river
        ]
    )

    game_api = api.GameApi(api_client)
    _ = await game_api.game_game_id_hand_post(
        game_id, GameNewHandDTO(deck=deck),
    )

    await bots[2].ante()
    await bots[3].ante()

    await bots[0].call()
    await bots[1].call()
    await bots[2].call()
    await bots[3].check()

    game_state = await bots[0].get_game_state()
    assert game_state.table.street[0].rank == help.CardRank.ACE.value
    assert game_state.table.street[0].suit == help.CardSuit.SPADE.value
    assert game_state.table.street[1].rank == help.CardRank.ACE.value
    assert game_state.table.street[1].suit == help.CardSuit.CLUB.value
    assert game_state.table.street[2].rank == help.CardRank.KING.value
    assert game_state.table.street[2].suit == help.CardSuit.CLUB.value


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
        game_id, GameNewHandDTO(deck=get_deck(42))
    )

    print(json.dumps(resp, indent=2))

    await bots[2].ante()
    await bots[3].ante()

    try:
        _ = await game_api.game_game_id_hand_post(
            game_id, GameNewHandDTO(deck=get_deck(42))
        )

    except api.ApiException as e:
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
        assert poker_error.type == "error"
        assert poker_error.data.error_code == 2025
