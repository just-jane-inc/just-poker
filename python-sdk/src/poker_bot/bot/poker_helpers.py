import logging
from enum import Enum

from openapi_client import (
    ApiClient,
    ApiException,
    Configuration,
    GameApi,
    GameCardDTO,
    GameGameDTO,
    GameNewGameConfigDTO,
    GamePostRequest,
    UserApi,
)
from poker_bot.bot.poker_exceptions import CustomException

logger = logging.getLogger("helper")


async def get_game_state(client: ApiClient, game_id: str) -> GameGameDTO | None:
    game_api = GameApi(client)
    resp = await game_api.game_game_id_state_get(game_id)
    return resp.data


async def start_game(client: ApiClient, game_id: str):
    game_api = GameApi(client)
    resp = await game_api.game_game_id_started_post(game_id)
    return resp.data


def create_connection(base_url: str, token: str = "") -> ApiClient:
    config = None
    if token:
        config = Configuration(host=base_url, access_token=token)
    else:
        config = Configuration(host=base_url)

    return ApiClient(config)


async def create_game(
    base_url: str,
    token: str,
    bb: int = 100,
    sb: int = 50,
    chips: dict[str, int] | None = None,
    player_count: int = 5,
    auto_start_hands: bool = True,
    denominations: list[int] | None = None,
) -> str | None:
    if not chips:
        chips = {"10": 10, "50": 5, "100": 2, "500": 1}

    if not denominations:
        denominations = [10, 50, 100, 500]
    conn = create_connection(base_url, token)
    api = GameApi(conn)
    dto = GameNewGameConfigDTO(
        big_blind=bb,
        small_blind=sb,
        starting_chips=chips,
        player_count=player_count,
        auto_starts_hands=auto_start_hands,
        chip_denominations=denominations,
    )

    resp = await api.game_post(GamePostRequest(dto))
    return resp.data


async def delete_user(base_url: str, token: str) -> str | None:
    api = UserApi(create_connection(base_url, token))
    try:
        resp = await api.user_me_delete()
        return None
    except ApiException as e:
        return e.body

    if resp.type == "error":
        print(resp.data)
        raise CustomException(resp.data)


class CardSuit(Enum):
    SPADE = ord("s")
    HEART = ord("h")
    DIAMOND = ord("d")
    CLUB = ord("c")
    UNKNOWN = ord("x")


class CardRank(Enum):
    ACE = ord("A")
    TWO = ord("2")
    THREE = ord("3")
    FOUR = ord("4")
    FIVE = ord("5")
    SIX = ord("6")
    SEVEN = ord("7")
    EIGHT = ord("8")
    NINE = ord("9")
    TEN = ord("T")
    JACK = ord("J")
    QUEEN = ord("Q")
    KING = ord("K")
    UNKNOWN = ord("x")


def get_unicode_mapping():
    mapping = dict()
    offset = 0x1F0A1 - 16
    for suit in CardSuit:
        if suit == CardSuit.UNKNOWN:
            continue

        offset += 16
        mapping[suit] = dict()
        for i, rank in enumerate(CardRank):
            if rank == CardRank.UNKNOWN:
                continue

            # operation wtf is the knight of spades
            if rank == CardRank.QUEEN:
                offset += 1

            mapping[suit][rank] = chr(offset + i)

        # like for real, the KNIGHT of spades?
        offset -= 1

    mapping[CardSuit.UNKNOWN] = dict()
    mapping[CardSuit.UNKNOWN][CardRank.UNKNOWN] = chr(0x1F0A0)
    return mapping


class Card:
    def __init__(self, rank: CardRank, suit: CardSuit):
        self._suit = suit
        self._rank = rank

    @property
    def Suit(self):
        return self._suit

    @property
    def Rank(self):
        return self._rank

    @classmethod
    def from_card_dto(cls, dto: GameCardDTO):
        rank = CardRank(dto.rank)
        suit = CardSuit(dto.suit)
        return cls(rank, suit)
