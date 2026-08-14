import argparse
import asyncio
import json
import os

from dotenv import load_dotenv

import openapi_client as api
import poker_bot.bot.poker_helpers as help

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
    required=False,
    help="the path to a file containing a configuration json that can be used jto create a game",
)

parser.add_argument(
    "--players",
    "--player-count",
    type=int,
    required=False,
    default=4,
    help="number of players that can join the game",
)

parser.add_argument("--bb", "--big-blind", type=int, required=False, default=100, help="the big blind to configure")

parser.add_argument("--sb", "--small-blind", type=int, required=False, default=50, help="the small blind to configure")


def load_config(file: str):
    with open(file, "r") as f:
        config = json.load(f)
        return api.GameNewGameConfigDTO.from_dict(config)


async def new_game(config: api.GameNewGameConfigDTO, token: str):
    resp = await help.create_game_from_config(base_url=base_url, config=config, token=token)
    print(resp)


def main():
    with open("config/api-token", "r") as f:
        token = str(f.read()).strip("\n")

    args = parser.parse_args()

    config: api.GameNewGameConfigDTO | None = None
    if args.config:
        config = load_config(args.config)
    else:
        config = api.GameNewGameConfigDTO(
            auto_starts_hands=True,
            big_blind=args.big_blind,
            small_blind=args.small_blind,
            chip_denominations=[10, 25, 50, 100, 500],
            player_count=args.player_count,
            starting_chips={"10": 10, "25": 5, "50": 5, "100": 5, "500": 1},
        )

    asyncio.run(new_game(config, token))


if __name__ == "__main__":
    main()
