import argparse
import json
import os
from datetime import datetime

import aiofiles
from dotenv import load_dotenv
from users.users import get_test_user

from poker_bot.bot import *
from poker_bot.bot import PokerBot

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
    "--id",
    type=str,
    required=True,
    help="the id of the game to going",
)

parser.add_argument(
    "--file",
    type=str,
    required=False,
    default="",
    help="the path to a file to write events to",
)

if os.name == "nt":
    # For windows only tool usage, fix for cert stuff
    try:
        import truststore

        truststore.inject_into_ssl()
    finally:
        pass


async def listen_as(user: str, game_id: str, file_path: str = ""):
    test_user = get_test_user(user)
    bot = PokerBot(base_url, test_user.token, test_user.user_id, game_id)
    print(test_user.token)

    now = datetime.now().strftime("%Y%m%d-%H%M%S")
    if file_path == "":
        file_path = f"listener_event_log-{user}-{game_id}-{now}.data"

    async with aiofiles.open(file_path, "a+") as f:
        await f.write(f"--- starting listener game:{game_id} user:{user} time:{now} ---\n\n")

        try:
            async for e in bot.events.stream():
                print(f"received event: [{e.id}] [{e.time_sent}] {e.event_type}")
                print(json.dumps(e._data, indent=2))
                await f.write(json.dumps(e.raw) + "\n")
        finally:
            await bot.stop_events()


def main():
    args = parser.parse_args()
    asyncio.run(listen_as(args.user, args.game_id))


if __name__ == "__main__":
    main()
