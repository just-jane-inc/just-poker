import argparse
import asyncio
import json
import os

from dotenv import load_dotenv

from poker_bot.bot.bot import *
from poker_bot.bot.bot import PokerBot
from poker_bot.tools.tui.setup_tui_example import get_test_user

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser("poker bot tool CLI")
parser.add_argument(
    "--user",
    type=str,
    required=True,
    help="the names of the test user to listen with",
)

parser.add_argument(
    "--game-id",
    type=str,
    required=True,
    help="the id of the game to going",
)


async def listen_as(user: str, game_id: str):
    test_user = get_test_user(user)
    bot = PokerBot(base_url, test_user.token, test_user.user_id, game_id)
    stream = bot.websocket_stream()
    await stream.connect()

    async for e in stream.events():
        print(f"received event: [{e.id}] [{e.time_sent}] {e.event_type}")
        print(json.dumps(e._data, indent=2))


def main():
    args = parser.parse_args()
    asyncio.run(listen_as(args.user, args.game_id))


if __name__ == "__main__":
    main()
