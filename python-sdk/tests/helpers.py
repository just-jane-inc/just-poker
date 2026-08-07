import os

from dotenv import load_dotenv

import poker_bot.tools.tui.setup_tui_example as examples

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")


def get_test_users() -> list[examples.TestUser]:
    test_users = examples.get_test_users()
    users = []
    for user in test_users:
        if user.username == "jill":
            continue
        users.append(user)

    return users
