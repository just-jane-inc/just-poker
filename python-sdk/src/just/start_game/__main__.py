import argparse
import asyncio
import os

from dotenv import load_dotenv

import poker_bot.poker_helpers as help

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")
token = os.getenv("TOKEN")

parser = argparse.ArgumentParser("poker bot tool CLI")

parser.add_argument(
    "--game-id",
    "--id",
    type=str,
    required=True,
    help="the id of the game to start",
)


async def start_game(game_id: str):
    conn = help.create_connection(base_url, token)
    r = await help.start_game(conn, game_id)
    print(r)


def main():
    args = parser.parse_args()
    asyncio.run(start_game(args.game_id))


if __name__ == "__main__":
    main()
