import pytest

import poker_bot.bot.poker_helpers as help
from openapi_client.models.game_round_type import GameRoundType
from poker_bot.bot import bot
from poker_bot.tools.tui.setup_tui_example import get_test_users

base_url = "http://localhost:7653"
# base_url = "https://game.bahms.org/api/poker"


# TODO: make some test users to re-use rather than making them each time
@pytest.mark.asyncio
async def test_all_check():
    users = get_test_users()
    assert len(users) == 4

    game_id = await help.create_game(base_url, str(users[0].token))
    assert game_id

    bots: list[bot.PokerBot] = []
    for user in users:
        b = bot.PokerBot(base_url, user.token, user.user_id, game_id)
        await b.join_game()
        bots.append(b)

    await bots[0].start_game()

    # setup
    await bots[2].send_action("ante", {"10": 5})
    await bots[3].send_action("ante", {"10": 10})

    # pre flop
    await bots[0].send_action("call", {"10": 10})
    await bots[1].send_action("call", {"10": 10})
    await bots[2].send_action("call", {"10": 5})
    await bots[3].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_flop

    # flop
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert state.table.current_round.current_round_type == GameRoundType.round_type_turn

    # turn
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})

    state = await bots[0].get_game_state()
    assert state is not None
    assert (
        state.table.current_round.current_round_type == GameRoundType.round_type_river
    )

    # river
    await bots[2].send_action("check", {})
    await bots[3].send_action("check", {})
    await bots[0].send_action("check", {})
    await bots[1].send_action("check", {})
