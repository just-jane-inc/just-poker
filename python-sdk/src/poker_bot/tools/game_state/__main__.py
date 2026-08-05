import argparse
import asyncio
import os

from dotenv import load_dotenv

from poker_bot.bot.poker_helpers import create_connection, get_game_state

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")

parser = argparse.ArgumentParser("poker bot tool CLI")
parser.add_argument(
    "--game-id",
    type=str,
    required=False,
    help="the id of the game to get the state of",
)


async def get_state(game_id: str):
    client = create_connection(
        base_url, "bahms.LqEwOpyhXZ7tZRUf.q_gZf2_hpcfdV3H6gvkEt7cRBb3Lcb-LNoYd-nGYE4oa"
    )

    resp = await get_game_state(client, game_id)
    print(resp.to_str())


def main():
    print("starting thing")
    with open("config/api-token", "r") as f:
        token = str(f.read()).strip("\n")
        print(token)

    args = parser.parse_args()
    asyncio.run(get_state(args.game_id))


if __name__ == "__main__":
    main()
