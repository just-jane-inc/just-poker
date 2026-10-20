import argparse
import asyncio
import json
import os

from dotenv import load_dotenv

import openapi_client as api
import poker_bot.poker_helpers as help

if os.name == "nt":
    # For windows only tool usage, fix for cert stuff
    try:
        import truststore

        truststore.inject_into_ssl()
    finally:
        pass

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser(
    "just__poker: create a game",
    usage="create-game --config config/new_game_config/0.json",
    description="""
    creates a game from a configuration json file that expresses a new game
    in the format described at python-sdk/src/openapi_client/docs/GameNewGameConfigDTO.md
    """,
)

parser.add_argument(
    "--config",
    type=str,
    required=True,
    help="the path to a file containing a configuration json that can be used jto create a game",
)


def load_config(file: str):
    with open(file, "r") as f:
        config = json.load(f)
        return api.GameNewGameConfigDTO.from_dict(config)


async def new_game(config: api.GameNewGameConfigDTO, token: str):
    assert base_url, "base api url not found check config/.env"
    resp = await help.create_game_from_config(base_url=base_url, config=config, token=token)
    print(resp)


def main():
    with open("config/api-token", "r") as f:
        token = str(f.read()).strip("\n")

    assert token, "token for creating game not found in config/api-token"
    args = parser.parse_args()

    config = load_config(args.config)
    assert config, "config not bound..."
    asyncio.run(new_game(config, token))


if __name__ == "__main__":
    main()
