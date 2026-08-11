import pytest
from helpers import (
    base_url,
    fill_deck_remainder,
    get_test_user,
    make_tests_work_for_fricking_windows,
)

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.models.game_card_dto import GameCardDTO
from openapi_client.models.game_game_id_hand_post_request import (
    GameGameIdHandPostRequest,
)
from openapi_client.models.game_new_hand_dto import GameNewHandDTO
from poker_bot.bot import bot

make_tests_work_for_fricking_windows()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_post_new_hand_with_deck():
    jill = get_test_user("jill")

    api_client = help.create_connection(base_url, jill.token)

    game_id = await help.create_game(
        base_url, jill.token, player_count=4, auto_start_hands=False
    )
    assert game_id

    wolf = get_test_user("wolf")
    wolf_bot = bot.PokerBot(base_url, wolf.token, wolf.user_id, game_id)
    await wolf_bot.join_game()

    red = get_test_user("red")
    red_bot = bot.PokerBot(base_url, red.token, red.user_id, game_id)
    await red_bot.join_game()

    rae = get_test_user("rae")
    rae_bot = bot.PokerBot(base_url, rae.token, rae.user_id, game_id)
    await rae_bot.join_game()

    jane = get_test_user("jane")
    jane_bot = bot.PokerBot(base_url, jane.token, jane.user_id, game_id)
    await jane_bot.join_game()

    await help.start_game(api_client, game_id)
    game_api = api.GameApi(api_client)
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

    resp = await game_api.game_game_id_hand_post(
        game_id, GameGameIdHandPostRequest(GameNewHandDTO(deck=deck))
    )

    assert resp

    await rae_bot.ante()
    await jane_bot.ante()

    await wolf_bot.fold()
    await red_bot.all_in()
    await rae_bot.fold()
    await jane_bot.all_in()

    game_state = await jane_bot.get_game_state()
    assert game_state.table.street[0].rank == help.CardRank.ACE.value
    assert game_state.table.street[0].suit == help.CardSuit.SPADE.value
    assert game_state.table.street[1].rank == help.CardRank.ACE.value
    assert game_state.table.street[1].suit == help.CardSuit.CLUB.value
    assert game_state.table.street[2].rank == help.CardRank.KING.value
    assert game_state.table.street[2].suit == help.CardSuit.CLUB.value

    deck = fill_deck_remainder(
        [
            GameCardDTO(rank=ord("A"), suit=ord("d")),  # rae
            GameCardDTO(rank=ord("A"), suit=ord("h")),  # jane
            GameCardDTO(rank=ord("2"), suit=ord("d")),  # wolf
            GameCardDTO(rank=ord("K"), suit=ord("h")),  # rae
            GameCardDTO(rank=ord("K"), suit=ord("d")),  # jane
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

    resp = await game_api.game_game_id_hand_post(
        game_id, GameGameIdHandPostRequest(GameNewHandDTO(deck=deck))
    )

    assert resp

    await jane_bot.ante()
    await wolf_bot.ante()
    await rae_bot.all_in()
    await jane_bot.all_in()
    await wolf_bot.all_in()

    state = await jane_bot.get_game_state()
    assert state

    assert help.chip_sum(state.table.players[0].stack) == 1550
    assert help.chip_sum(state.table.players[3].stack) == 2650
