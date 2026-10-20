from dataclasses import dataclass

import poker_bot.poker_helpers as help
from poker_bot.bot import PokerBot


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
    with open("config/test_users.csv", "r") as f:
        for user in f:
            user = user.rstrip()
            username, userid, token = user.split(",")
            users.append(TestUser(userid, username, token))

    return users


async def get_test_setup(base_url: str, player_count: int) -> TestSetup:
    """gets the default basic test setup for provided number of players

    will create a game with the @player_count players populated with users acquired
    from config/test_users.csv. the game will be started before being returned.
    """
    bots: dict[str, TestBot] = {}
    users = get_test_users()

    game_id = await help.create_game(base_url, users[0].token, player_count=player_count)
    assert game_id

    some_bot: PokerBot | None = None
    for user in get_test_users()[:player_count]:
        if user.username == "jill":
            continue

        bot = PokerBot(base_url, user.token, user.user_id, game_id)
        if not some_bot:
            some_bot = bot

        await bot.join_game()
        bots[user.username] = TestBot(user, bot)

    assert some_bot, "no bots created, is config/test_users.csv empty?"
    await some_bot.start_game()

    return TestSetup(bots, game_id)


async def setup(base_url: str, player_count: int = 4):
    setup = await get_test_setup(base_url, player_count)
    print(f"game_id: {setup.game_id}")
    for id, user in setup.bots.items():
        print(f"user_name: {id}")
        print(f"user_id: {user.user.user_id}")
