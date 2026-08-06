import argparse
import asyncio
import os

from dotenv import load_dotenv

from poker_bot.bot.poker_helpers import create_game

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser("poker bot tool CLI")
parser.add_argument(
    "--players",
    type=int,
    required=False,
    default=4,
    help="number of players that can join the game",
)
parser.add_argument(
    "--bb", type=int, required=False, default=100, help="the big blind to configure"
)
parser.add_argument(
    "--sb", type=int, required=False, default=50, help="the small blind to configure"
)


async def new_game(player_count: int, big_blind: int, small_blind: int, token: str):
    resp = await create_game(
        base_url, token, big_blind, small_blind, player_count=player_count
    )
    print(resp)


def main():
    print("starting thing")
    with open("config/api-token", "r") as f:
        token = str(f.read()).strip("\n")
        print(token)

    args = parser.parse_args()
    asyncio.run(new_game(args.players, args.bb, args.sb, token))


if __name__ == "__main__":
    main()
