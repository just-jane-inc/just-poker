import argparse
import asyncio

from src.poker_helpers import create_game

parser = argparse.ArgumentParser("poker bot tool CLI")
parser.add_argument("--players", type=int, required=False, default=4, help="number of players that can join the game")
parser.add_argument("--bb", type=int, required=False, default=100, help="the big blind to configure")
parser.add_argument("--sb", type=int, required=False, default=50, help="the small blind to configure")
parser.add_argument("--local", action="store_true", default=False, help="the small blind to configure")


async def main(player_count: int, big_blind: int, small_blind: int, local: bool, token: str):
    base_url = "http://localhost:7653" if local else "https://game.bahms.org/api/poker"
    print(base_url)
    resp = await create_game(base_url, token, big_blind, small_blind, player_count=player_count)
    print(resp)


if __name__ == "__main__":
    print("starting thing")
    with open("tools/api-token", "r") as f:
        token = str(f.read()).strip("\n")
        print(token)

    args = parser.parse_args()
    asyncio.run(main(args.players, args.bb, args.sb, args.local, token))
