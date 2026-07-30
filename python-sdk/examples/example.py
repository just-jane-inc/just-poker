import asyncio
import json
import time
from dataclasses import dataclass

from src.bot import *
from src.poker_exceptions import *
from src.poker_helpers import *

base_url = "http://localhost:7653"


@dataclass
class User:
    Name: str
    Token: str
    UserID: str


users: list[User] = []


async def main():
    for i in range(5):
        name = f"test-user-jane-{i}"
        user = await create_user(base_url, "test_twitch_id_jane", name)
        assert user
        users.append(User(name, str(user.token), str(user.user_id)))

    game_id = await create_game(base_url, users[0].Token)
    assert game_id

    bots: list[PokerBot] = []
    for user in users:
        bot = PokerBot(base_url, user.Token, int(game_id))
        bots.append(bot)
        await bot.join_game()

    await bots[0].start_game()

    await bots[2].send_action("ante", {"50": 1})
    await bots[3].send_action("ante", {"50": 2})

    state = await bots[0].get_game_state()
    assert state

    print(json.dumps(state.dict(), indent=2))


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except Exception:
        # logger.error(f"encountered exception: {e}")
        for user in users:
            asyncio.run(delete_user(base_url, user.Token))

        raise
