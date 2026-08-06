import argparse
import asyncio
import os

from dotenv import load_dotenv

from poker_bot.bot.bot import PokerBot
from poker_bot.tools.tui.setup_tui_example import get_test_users

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser("poker bot tool CLI")
parser.add_argument(
    "--users",
    type=str,
    required=True,
    nargs="+",
    help="the names of test users to join with",
)

parser.add_argument(
    "--game-id",
    type=str,
    required=True,
    help="the id of the game to going",
)


async def join_game_as(users: list[str], game_id: str):
    print(f"joining as {users}")
    test_users = get_test_users()
    for user in test_users:
        if user.username not in users:
            continue

        bot = PokerBot(base_url, user.token, user.user_id, game_id)
        await bot.join_game()


def main():
    args = parser.parse_args()
    asyncio.run(join_game_as(args.users, args.game_id))


if __name__ == "__main__":
    main()
