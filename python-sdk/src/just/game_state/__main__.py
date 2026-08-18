import argparse
import asyncio
import os

import aiofiles
from dotenv import load_dotenv
from httpx import ConnectError

import poker_bot.poker_helpers as help
from openapi_client import ApiException

parser = argparse.ArgumentParser(
    description="gets game state for a game by id",
)

parser.add_argument(
    "--id",
    "--game-id",
    type=str,
    required=True,
    help="the id of the game to get the state of",
)

parser.add_argument(
    "--url",
    type=str,
    required=False,
    help="an override for the base_url to use",
)

parser.add_argument(
    "--token-file",
    type=str,
    required=False,
    help="an override for the filepath containing an api token to use",
)


async def get_state(game_id: str, token_file: str, base_url: str):
    async with aiofiles.open(token_file, "r") as f:
        token = await f.read()
        token = str(token.strip("\n"))

    client = help.create_connection(base_url, token)
    try:
        resp = await help.get_game_state(client, game_id)
        print(resp.to_str())
    except ApiException as e:
        print(f"encountered exception when getting game state: {e}")
    except ConnectError:
        print(f"connection could not be established - is the server at {base_url} online?")


def main():
    base_url = ""
    token_file = ""

    args = parser.parse_args()
    if args.url:
        base_url = args.url
    else:
        load_dotenv("config/.env")
        base_url = os.getenv("BASE_URL")

    if args.token_file:
        token_file = args.token_file
    else:
        token_file = "config/api-token"

    asyncio.run(get_state(args.id, token_file, base_url))


if __name__ == "__main__":
    main()
