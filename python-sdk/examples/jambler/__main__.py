import argparse
import asyncio
import json
import logging

import poker_bot.bot.poker_helpers as help
from poker_bot.bot.bot import PokerBot

import openapi_client as api
from poker_bot.bot import event_hub

parser = argparse.ArgumentParser(prog="just__poker: call bot", description="an example bot that always calls")
parser.add_argument("--game-id", "--id", type=str, required=True, help="the id of the game to join")
parser.add_argument("--config", type=str, required=True, help="the user configuration file to load with")
parser.add_argument("--url", type=str, required=True, help="the base url for the game server")
logger = logging.getLogger("jambler")


def main():
    args = parser.parse_args()
    bot: PokerBot | None = None
    with open(args.config, "r") as f:
        logger.debug("loading configuration file")
        config = json.load(f)
        logger.debug(f"config:\n {json.dumps(config, indent=2)}")
        bot = PokerBot(base_url=args.url, token=config["token"], user_id=config["user_id"], game_id=args.game_id)
        logger.debug("bot created")

    asyncio.run(run_jambler_bot(bot))


async def run_jambler_bot(bot: PokerBot):
    async with bot:
        await bot.join_game()
        game_over = asyncio.Event()
        game_started = asyncio.Event()
        game_api = bot.get_game_api()

        async def on_game_over(e: event_hub.Event):
            logger.info("game over!")
            game_over.set()

        async def on_game_started(e: event_hub.Event):
            logger.info("game started!")
            game_started.set()

        async def on_game_state_changed(e: event_hub.Event):
            logger.info("processing game state")
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

            game_state: api.GameGameDTO = e.data
            if not bot.is_my_turn():
                return

            try:
                # it is your turn and you have to procude an action, first check if the action must be an ante
                if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypeAnte:
                    await bot.ante()
                    return

                if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypePreFlop:
                    total = sum((int(denom) * count for denom, count in bot._player.current_bet.items()))
                    if total == game_state.table.current_round.bet:
                        await bot.check()
                    elif bot.chip_total() > game_state.table.current_round.bet * 4:
                        await bot.call()
                    else:
                        await bot.fold()

                    return

                if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypeFlop:
                    flop = game_state.table.street
                    cards = [bot._player.hole[0], bot._player.hole[1], flop[0], flop[1], flop[2]]

                    request = api.HandEvaluatorEvaluatePostRequest(cards)
                    resp = await game_api.hand_evaluator_evaluate_post(request)

                    if resp.error != "":
                        print("big yikes")
                        await bot.all_in()
                    else:
                        if resp.evaluation < 2130:
                            await bot.all_in()
                        else:
                            await bot.fold()
                    return

                logger.error("something is wrong")

            except api.ApiException as ex:
                err: api.JustErrorDTO | None = help.get_error_from_exception(ex)
                logger.error(f"encountered error while preforming action [{err.error_code}]: {err.error}")

        logger.info("subscribing")
        subscribers = [
            bot.events.subscribe(event_hub.EventType.GAME_ENDING, on_game_over),
            bot.events.subscribe(event_hub.EventType.STARTING_GAME, on_game_started),
            bot.events.subscribe(event_hub.EventType.GAME_STATE_UPDATE, on_game_state_changed),
        ]

        logger.info("getting game state")
        state = await bot.get_game_state()
        if state.started_at is None:
            logger.info("game has not started, waiting for it")
            await asyncio.wait_for(game_started.wait(), None)

        logger.info("everything is okay")
        await asyncio.wait_for(game_over.wait(), None)

        for subscriber in subscribers:
            subscriber.unsubscribe()


if __name__ == "__main__":
    main()
