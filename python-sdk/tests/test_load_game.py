import json

import aiofiles
import helpers
import pytest

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.api.game_api import GameApi

helpers.make_tests_work_for_fricking_windows()


@pytest.mark.asyncio
async def test_load_game_from_state():
    jill = helpers.get_test_user("jill")

    api_client = help.create_connection(helpers.base_url, jill.token)

    game_id = await help.create_game(helpers.base_url, jill.token, auto_start_hands=False)
    assert game_id

    game_api: api.GameApi = GameApi(api_client)

    async with aiofiles.open("tests/data/load_game_test_data_1.json", "r") as f:
        json_str = await f.read()
        data = json.loads(json_str)

    table_dto: api.GameTableDTO = api.GameTableDTO.from_dict(data)
    assert table_dto is not None
    # TODO
    # await game_api.game_game_id_state_post(game_id, table_dto)
