from enum import Enum

import openapi_client as poker_api

from typing import Dict

import logging

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger("testing")


class CustomException(Exception):
    pass


def create_game(
    base_url: str,
    token: str,
    bb: int = 100,
    sb: int = 50,
    chips: Dict[str, int] | None = None,
) -> str | None:
    if not chips:
        chips = {"10": 10, "50": 5, "100": 2, "500": 1}
    conn = create_connection(base_url, token)
    api = poker_api.GameApi(conn)
    dto = poker_api.GameNewGameConfigDTO(
        big_blind=bb,
        small_blind=sb,
        starting_chips=chips,
        player_count=5,
    )

    try:
        resp = api.game_post(poker_api.GamePostRequest(dto))
        return resp.data
    except poker_api.ApiException as e:
        logger.error(f"encountered error creating game: {e}")


def create_connection(base_url: str, token: str = "") -> poker_api.ApiClient:
    config = None
    if token:
        config = poker_api.Configuration(host=base_url, access_token=token)
    else:
        config = poker_api.Configuration(host=base_url)

    return poker_api.ApiClient(config)


def create_user(
    base_url: str, twitch_id: str, display_name: str
) -> poker_api.UserApiKey | None:
    api = poker_api.UserApi(create_connection(base_url))
    dto = poker_api.UserUserDTO(
        display_name=display_name, twitch_id=twitch_id, user_type="test-user"
    )
    resp = api.user_post(poker_api.UserPostRequest(dto))
    return resp.data


def delete_user(base_url: str, token: str) -> str | None:
    api = poker_api.UserApi(create_connection(base_url, token))
    try:
        resp = api.user_me_delete()
        return None
    except poker_api.ApiException as e:
        return e.body

    if resp.type == "error":
        print(resp.data)
        raise CustomException(resp.data)


class CardSuit(Enum):
    SPADE = ord("s")
    CLUB = ord("c")
    HEART = ord("h")
    DIAMOND = ord("d")


class CardRank(Enum):
    ACE = ord("A")
    KING = ord("K")
    QUEEN = ord("Q")
    JACK = ord("J")
    TEN = ord("T")
    NINE = ord("9")
    EIGHT = ord("8")
    SEVEN = ord("7")
    SIX = ord("6")
    FIVE = ord("5")
    FOUR = ord("4")
    THREE = ord("3")
    TWO = ord("2")


class PokerBot:
    def __init__(self, base_url, token, game_id):
        """A poker bot

        Args:
            base_url: the url for the poker server to connect to
            token: the user authorization token to use for this bot
            game_id: the id of the game to interact with

        Raises:
            CustomException: if an argument is not provided
        """
        if not token:
            raise CustomException("api_key not provided")

        if not base_url:
            raise CustomException("base url not provided")

        if not game_id:
            raise CustomException("base url not provided")

        self._game_id = game_id
        self._api_client = create_connection(base_url, token)
        self._user_api = poker_api.UserApi(self._api_client)
        self._game_api = poker_api.GameApi(self._api_client)
        self._joined = False

    def join_game(self):
        if self._joined:
            return

        _ = self._game_api.game_game_id_player_post(self._game_id)

    def start_game(self):
        self._game_api.game_game_id_started_post(self._game_id)

    def get_game_state(self):
        resp = self._game_api.game_game_id_state_get(self._game_id)
        return resp.data

    def send_action(self, intent: str, bet: Dict[str, int] | None):
        dto = poker_api.GamePlayerActionDTO(chips=bet, intent=intent)
        req = poker_api.GameGameIdActionPostRequest(dto)
        self._game_api.game_game_id_action_post(self._game_id, req)
