import asyncio
import logging

import httpx

import src.poker_exceptions as ex
import src.poker_helpers as help
from openapi_client import *

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger("bot")


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
            raise ex.CustomException("api_key not provided")

        if not base_url:
            raise ex.CustomException("base url not provided")

        if not game_id:
            raise ex.CustomException("base url not provided")

        self._game_id = game_id
        self._api_client = help.create_connection(base_url, token)
        self._user_api = UserApi(self._api_client)
        self._game_api = GameApi(self._api_client)
        self._joined = False

    async def join_game(self):
        if self._joined:
            return

        _ = await self._game_api.game_game_id_player_post(self._game_id)

    async def start_game(self):
        await self._game_api.game_game_id_started_post(self._game_id)

    async def get_game_state(self) -> GameGameDTO | None:
        resp = await self._game_api.game_game_id_state_get(self._game_id)
        return resp.data

    async def send_action(self, intent: str, bet: dict[str, int] | None):
        dto = GamePlayerActionDTO(chips=bet, intent=intent)
        req = GameGameIdActionPostRequest(dto)
        await self._game_api.game_game_id_action_post(self._game_id, req)
