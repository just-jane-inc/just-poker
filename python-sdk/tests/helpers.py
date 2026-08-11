import os
import random

from dotenv import load_dotenv

import poker_bot.tools.tui.setup_tui_example as examples
from openapi_client.models.game_card_dto import GameCardDTO

load_dotenv("config/.env")
base_url = os.getenv("BASE_URL")


def make_tests_work_for_fricking_windows():
    if os.name == "nt":
        try:
            import truststore

            truststore.inject_into_ssl()
        finally:
            pass


def get_deck(seed: int | None = None) -> list[GameCardDTO]:
    ranks = ["2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"]
    suits = ["s", "h", "c", "d"]
    deck = [GameCardDTO(rank=ord(r), suit=ord(s)) for r in ranks for s in suits]

    if seed is not None:
        random.seed(seed)

    random.shuffle(deck)
    return deck


def fill_deck_remainder(deck: list[GameCardDTO]) -> list[GameCardDTO]:
    new_deck = deck[:]
    taken_cards = set()
    for card in new_deck:
        taken_cards.add((card.rank, card.suit))

    random_deck = get_deck()
    for card in random_deck:
        if (card.rank, card.suit) in taken_cards:
            continue

        taken_cards.add((card.rank, card.suit))
        new_deck.append(card)

    return new_deck


def get_test_users() -> list[examples.TestUser]:
    test_users = examples.get_test_users()
    users = []
    for user in test_users:
        if user.username == "jill":
            continue
        users.append(user)

    return users


def get_test_user(username: str) -> examples.TestUser | None:
    users = examples.get_test_users()
    for user in users:
        if user.username == username:
            return user

    return None
