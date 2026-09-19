import pytest
from helpers import (
    base_url,
    get_test_user,
    get_test_users,
    make_tests_work_for_fricking_windows,
)

import openapi_client as api
import poker_bot.poker_helpers as help
from openapi_client.api.game_api import GameApi
from openapi_client.models.game_player_action_dto import GamePlayerActionDTO
from openapi_client.models.game_player_intent import GamePlayerIntent
from poker_bot import bot

make_tests_work_for_fricking_windows()


@pytest.mark.asyncio()
@pytest.mark.timeout(30)
async def test_action_errors():
    ab_viney = get_test_user("ab_viney")

    api_client = help.create_connection(base_url, ab_viney.token)

    game_id = await help.create_game(base_url, ab_viney.token)
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

    # bot 2 must ante the small blind which is currently at 50,
    # a bet of 100 is too much
    await game_api.game_game_id_action_post(
        game_id=game_id,
        game_player_action_dto=GamePlayerActionDTO(
            chips={"50": 1}, intent=GamePlayerIntent.PlayerIntentAnte, player_id=bots[2]._user_id
        ),
    )

    try:
        thing = bots[2].get_game_api()
        await thing.game_game_id_action_post(
            game_id=game_id,
            game_player_action_dto=GamePlayerActionDTO(
                chips={"100": 1}, intent=GamePlayerIntent.PlayerIntentAnte, player_id=bots[3]._user_id
            ),
        )

        assert False
    except api.ApiException as e:
        assert True
        assert e.status == 403
        poker_error = api.JustResponseMessageJustErrorDTO.from_json(e.body)
        assert poker_error.type == "error"
