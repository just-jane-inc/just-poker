import helpers as test_helpers
import pytest

import openapi_client
import poker_bot.poker_helpers as help
from openapi_client.models.game_card_dto import GameCardDTO

test_helpers.make_tests_work_for_fricking_windows()


@pytest.mark.asyncio
@pytest.mark.timeout(30)
async def test_chip_exchange_happens():
    jane = test_helpers.get_test_user("jane")
    api = openapi_client.GameApi(help.create_connection(test_helpers.base_url, jane.token))
    eval = await api.hand_evaluator_evaluate_post(
        game_card_dto=[
            GameCardDTO(rank=help.CardRank.ACE.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.KING.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.QUEEN.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.JACK.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.TEN.value, suit=help.CardSuit.SPADE.value),
            GameCardDTO(rank=help.CardRank.SIX.value, suit=help.CardSuit.HEART.value),
            GameCardDTO(rank=help.CardRank.SEVEN.value, suit=help.CardSuit.DIAMOND.value),
        ]
    )

    assert eval.evaluation == 1
