from dataclasses import dataclass

from src.bot import PokerBot
import src.poker_helpers as help
import asyncio


@dataclass
class TestUser:
    user_id: str
    username: str
    token: str


@dataclass
class TestBot:
    user: TestUser
    bot: PokerBot


@dataclass
class TestSetup:
    bots: dict[str, TestBot]
    game_id: str


def get_test_user(username: str) -> TestUser | None:
    users = get_test_users()
    for user in users:
        if user.username == username:
            return user
    return None


def get_test_users() -> list[TestUser]:
    users: list[TestUser] = []
    with open("examples/test_users.csv", "r") as f:
        for user in f:
            user = user.rstrip()
            username, userid, token = user.split(",")
            users.append(TestUser(userid, username, token))

    return users


async def get_test_setup(base_url: str) -> TestSetup:
    bots: dict[str, TestBot] = dict()
    users = get_test_users()

    game_id = await help.create_game(base_url, users[0].token, player_count=6)
    assert game_id

    for user in get_test_users():
        bot = PokerBot(base_url, user.token, user.user_id, game_id)
        await bot.join_game()
        return None
        bots[user.username] = TestBot(user, bot)

    #    await bots["jane"].bot.start_game()

    return TestSetup(bots, game_id)


async def setup(base_url: str):
    setup = await get_test_setup(base_url)
    print(f"game_id: {setup.game_id}")
    for id, user in setup.bots.items():
        print(f"user_name: {id}")
        print(f"user_id: {user.user.user_id}")
