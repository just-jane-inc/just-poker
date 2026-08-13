import argparse
import asyncio
import json
import logging

import openapi_client as api
import poker_bot.bot.poker_helpers as help
from openapi_client.models.game_game_dto import GameGameDTO
from poker_bot.bot import event_hub
from poker_bot.bot.bot import PokerBot

parser = argparse.ArgumentParser(prog="just__poker: call bot", description="an example bot that always calls")

parser.add_argument("--id", "--game-id", type=str, required=True, help="the id of the game to join")

logger = logging.getLogger("call_bot")


def main():
    args = parser.parse_args()
    bot: PokerBot | None = None
    with open("examples/call_bot/config", "r") as f:
        logger.debug("loading configuration file")
        config = json.load(f)
        logger.debug(f"config:\n {json.dumps(config, indent=2)}")
        bot = PokerBot(base_url=config["url"], token=config["token"], user_id=config["user_id"], game_id=args.game_id)
        logger.debug("bot created")

    asyncio.run(run_all_in_bot(bot))


async def run_all_in_bot(bot: PokerBot):
    async with bot:
        await bot.join_game()
        game_over = asyncio.Event()
        game_started = asyncio.Event()

        async def on_game_over(e: event_hub.Event):
            game_over.set()

        async def on_game_started(e: event_hub.Event):
            game_started.set()

        async def on_game_state_changed(e: event_hub.Event):
            if e.event_type is not event_hub.EventType.GAME_STATE_UPDATE:
                logger.error(f"received event type {e.event_type} in error on game_state subscriber")
                return

            if e.data is None:
                return

            if not isinstance(e.data, api.GameGameDTO):
                logger.error(
                    f"received data that does not match the expected type: {type(e.data)} != {type(api.GameGameDTO)}"
                )

                return

            game_state: GameGameDTO = e.data
            if not bot.is_my_turn():
                return

            try:
                if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypeAnte:
                    await bot.ante()
                else:
                    await bot.all_in()
            except api.ApiException as ex:
                err: api.JustErrorDTO | None = help.get_error_from_exception(ex)
                logger.error(f"encountered error while preforming action [{err.error_code}]: {err.error}")

        subscribers = [
            bot.events.subscribe(event_hub.EventType.GAME_OVER, on_game_over),
            bot.events.subscribe(event_hub.EventType.STARTING_GAME, on_game_started),
            bot.events.subscribe(event_hub.EventType.GAME_STATE_UPDATE, on_game_state_changed),
        ]

        state = await bot.get_game_state()
        if state.started_at is None:
            await asyncio.wait_for(game_started.wait())

        await asyncio.wait_for(game_over.wait())

        for subscriber in subscribers:
            subscriber.unsubscribe()


if __name__ == "__main__":
    main()
