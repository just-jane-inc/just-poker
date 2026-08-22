import argparse
import asyncio
import os

from dotenv import load_dotenv

import openapi_client as api
import poker_bot.poker_helpers as help
from poker_bot.poker_exceptions import CustomException

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")
token = os.getenv("TOKEN")

parser = argparse.ArgumentParser("test a game state")

parser.add_argument(
    "--file-path",
    "-f",
    type=str,
    required=True,
    help="the path to the file to test",
)


async def test_game(dto: api.GameGameDTO):
    conn = help.create_connection(base_url, token)
    game_id = await help.create_game_from_config(base_url=base_url, token=token, config=dto.game_config)
    assert game_id

    game_api = api.GameApi(conn)
    await game_api.game_game_id_state_post(game_id, dto.table)


def main():
    args = parser.parse_args()

    if not args.file_path:
        raise CustomException("yawn")

    with open(args.file_path, "r") as f:
        game_dto = api.GameGameDTO.from_json(f.read())

    asyncio.run(test_game(game_dto))


if __name__ == "__main__":
    main()
