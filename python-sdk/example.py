from dataclasses import dataclass

from thing import *

import json
import time


@dataclass
class User:
    Name: str
    Token: str
    UserID: str


base_url = "http://localhost:7653"
users: list[User] = []


def main():
    for i in range(5):
        name = f"test-user-jane-{i}"
        user = create_user(base_url, "test_twitch_id_jane", name)
        assert user
        users.append(User(name, str(user.token), str(user.user_id)))

    game_id = create_game(base_url, users[0].Token)
    assert game_id

    bots: list[PokerBot] = []
    for user in users:
        bot = PokerBot(base_url, user.Token, int(game_id))
        bots.append(bot)
        bot.join_game()

    bots[0].start_game()

    bots[2].send_action("ante", {"50": 1})
    bots[3].send_action("ante", {"50": 2})

    state = bots[0].get_game_state()
    assert state

    print(json.dumps(state.dict(), indent=2))


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        logger.error(f"encountered exception: {e}")
        for user in users:
            delete_user(base_url, user.Token)
