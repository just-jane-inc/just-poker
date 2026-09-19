import argparse
import asyncio
import json
import logging

import openapi_client as api
import poker_bot.bot as pb
import poker_bot.poker_helpers as help
from poker_bot import event_hub as eh

parser = argparse.ArgumentParser(prog="just__poker: call bot", description="an example bot that always calls")
parser.add_argument("--game-id", "--id", type=str, required=True, help="the id of the game to join")
parser.add_argument("--config", type=str, required=True, help="the user configuration file to load with")
parser.add_argument(
    "--url",
    type=str,
    required=False,
    default="https://game.bahms.org/api/poker",
    help="the base url for the game server",
)
logger = logging.getLogger("jambler")


def main():
    args = parser.parse_args()
    bot: pb.PokerBot | None = None
    with open(args.config, "r") as f:
        logger.debug("loading configuration file")
        config = json.load(f)
        logger.debug(f"config:\n {json.dumps(config, indent=2)}")
        bot = pb.PokerBot(base_url=args.url, token=config["token"], user_id=config["user_id"], game_id=args.game_id)
        logger.debug("bot created")

    asyncio.run(run_jambler_bot(bot))


async def execute_bot_turn(bot: pb.PokerBot, game_state: api.GameGameDTO):
    """execute a turn for the bot

    Args:
        bot: the bot to make a move for
        game_state: the state of the game to react to

    Think this code needs comments ? or that the comments here are useless?
    well come to twitch and add your own and help us solve this problem.....
    or make it worse - Goblinz181
    """

    # the game_api here allows you to access the api client
    # that the bot uses to talk to the server, allowing you to
    # use the full api without the PokerBot layer
    game_api = bot.get_game_api()

    try:
        # it is your turn and you have to produce an action, first check if the action must be an ante
        if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypeAnte:
            # this is really the only valid action here, note that the PokerBot SDK is handling the
            # heavy lifing of determining which chips you ante with
            await bot.ante()
            return

        if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypePreFlop:
            # we compute the total here in order to track how many chips the bot has already
            # produced in this betting round. we can use this to determine when we can
            # check, and preforming things like "raise to X"
            total = sum((int(denom) * count for denom, count in bot._player.current_bet.items()))

            # this bot just tries to get through the pre-flop and see the flop
            # if we can check we do it every single time
            if total == game_state.table.current_round.bet:
                await bot.check()

            # if our total stack is at least 4x the current bet we will just call
            # this is a terrible strategy, its an example bot...just you know...
            # rest of the owl is on you
            elif bot.chip_total() > game_state.table.current_round.bet * 4:
                await bot.call()
            else:
                await bot.fold()

            return

        if game_state.table.current_round.current_round_type == api.GameRoundType.RoundTypeFlop:
            # now that we have seen the flop we are able to evaluate our position
            # using a provided endpoint
            flop = game_state.table.street
            cards = [bot._player.hole[0], bot._player.hole[1], flop[0], flop[1], flop[2]]

            # we make a request that includes our hand (our hole cards as well as the flop)
            # to see its strength
            resp = await game_api.hand_evaluator_evaluate_post(cards)

            # if we have an error...idk just go all in? who cares
            if resp.error != "":
                print("big yikes")
                await bot.all_in()
            else:
                # the evaluation provided by the endpoint rates all cards using
                # a cactus-kev algorithm. 1 is a royal flush (the best hand) as numbers
                # increase the hand is worse and worse (~7500 is the worst hand)
                #
                # if we have a decent hand we just go all in
                if resp.evaluation < 2130:
                    await bot.all_in()
                # otherwise we fold
                else:
                    await bot.fold()
            return

        logger.error("something is wrong, we should have went all in or folded...how can it be my turn?")
    except api.ApiException as ex:
        err: api.JustErrorDTO | None = help.get_error_from_exception(ex)
        logger.error(f"encountered error while preforming action [{err.error_code}]: {err.error}")


async def run_jambler_bot(bot: pb.PokerBot):
    # we create an async context for the bot
    async with bot:
        # try to join the provided game, if it fails
        # we assume the game has already started
        try:
            print("joining game")
            await bot.join_game()
        except Exception as e:
            logger.warning(f"assuming the game has already started - right? {e}")

        # we use these events to block until
        # specific state is achieved
        game_over = asyncio.Event()
        game_started = asyncio.Event()

        async def on_game_over(e: eh.Event):
            logger.info("game over!")
            game_over.set()

        async def on_game_started(e: eh.Event):
            logger.info("game started!")
            game_started.set()

        async def on_game_state_changed(e: eh.Event):
            logger.info("processing game state")
            if e.event_type is not eh.EventType.GAME_STATE_UPDATE:
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

            await execute_bot_turn(bot, game_state)

        logger.info("subscribing")
        subscribers = [
            bot.events.subscribe(eh.EventType.GAME_ENDING, on_game_over),
            bot.events.subscribe(eh.EventType.STARTING_GAME, on_game_started),
            bot.events.subscribe(eh.EventType.GAME_STATE_UPDATE, on_game_state_changed),
        ]

        logger.info("getting game state")
        state = await bot.get_game_state()

        if state.started_at is None:
            logger.info("game has not started, waiting for it")
            await asyncio.wait_for(game_started.wait(), None)
        elif bot.is_my_turn():
            await execute_bot_turn(bot, state)

        logger.info("everything is okay")
        await asyncio.wait_for(game_over.wait(), None)

        for subscriber in subscribers:
            subscriber.unsubscribe()


if __name__ == "__main__":
    main()
