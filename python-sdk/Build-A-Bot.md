Looking to start building a bot? Here is the simplest
setup once you have the SDK imported and functional.

## Bot Object

First, create a bot object. This will be your access to the whole game.

Store it somewhere accessible from your code anywhere you need it.

This bot is given a game_id, so this instance will only have access to this one game.

```py 
    from poker_bot.bot.bot import PokerBot
    # ...

    bot = PokerBot(base_url = "https://game.bahms.org/api/poker",  
                   token    = "",                  # Your Bot Token
                   user_id  = "",                  # Your Bot User Id, I.E., 42
                   game_id  = ""                   # Game Id, I.E., "4201"
          )

    await bot.start_events()                       # Guarantee the bot will start listening for events.
    await bot.join_game()                          # Join the game you are configured for 
```

Reminder, a user token can only be connected *once* to a Game's events websocket at a time.

## Subscriptions

```py
    from poker_bot.bot import event_hub
    import openapi_client as api
    # ...

    async def on_game_state_changed(e: event_hub.Event):
        # React to game states. This is what gives you everything you want
        if e.data is None or not isinstance(e.data, api.GameGameDTO):
            return
        pass
    
    # Just a list of subscriptions to make collecting them easier
    subscribers = [
        # You can subscribe to a single event type and a callback function as this
        bot.events.subscribe(event_hub.EventType.GAME_ENDING, on_game_over),
        
        # Or multiple at once to a single callback
        bot.events.subscribe([
            event_hub.EventType.GAME_STATE_UPDATE,
            event_hub.EventType.WELCOME,
            event_hub.EventType.STARTING_GAME
        ], on_game_state_changed),
        # ...
    ]
    
    # ... When closing the bot or disconnecting unsubscribe
    
    for subscriber in subscribers:
        subscriber.unsubscribe()

```

## Bot Lifecycle

Here is a fuller example showing how, given a game id, you can:

1) Create a Bot
2) Subscribe to essential events
3) Wait for events before acting, blocking code
4) Cleanup

This example is similar to what is used in the `examples/jambler/__main__.py`

```py
import asyncio

from poker_bot.bot.bot import PokerBot
from poker_bot.bot import event_hub
import openapi_client as api


def main():
    bot = PokerBot(base_url = "https://game.bahms.org/api/poker",  
               token    = "",                  # Your Bot Token
               user_id  = "",                  # Your Bot User Id, I.E., 42
               game_id  = ""                   # Game Id, I.E., "4201"
    )
    asyncio.run(run_bot(bot))
    
async def run_bot(bot: PokerBot): 
    await bot.start_events()
    await bot.join_game()
    
    # These are used to help the program "wait"/block until we get info on the game
    game_over = asyncio.Event()
    game_started = asyncio.Event()
    
    async def on_game_over(e: event_hub.Event):
        game_over.set()
    
    async def on_game_started(e: event_hub.Event):
        game_started.set()
    
    async def on_game_state_changed(e: event_hub.Event):
        if e.data is None or not isinstance(e.data, api.GameGameDTO):
            return
        
        game_state: api.GameGameDTO = e.data
        # ... Logic for analyzing hand, stack, cards, etc etc.
        # Make your Moves too... or something, kicks rocks...
        # The PokerBot will handle updating and give you access to the current stack, hand, hole, river, etc. 
        # ...
    
    # Subscribe
    subscribers = [
            bot.events.subscribe(event_hub.EventType.GAME_ENDING, on_game_over),
            bot.events.subscribe(event_hub.EventType.STARTING_GAME, on_game_started), 
            bot.events.subscribe([
                event_hub.EventType.GAME_STATE_UPDATE,
                event_hub.EventType.WELCOME,
                event_hub.EventType.STARTING_GAME
            ], on_game_state_changed)
    ]
        
    state = await bot.get_game_state()
    if state.started_at is None:
        # Wait for the game_started event to be received (calls `on_game_started`)
        await asyncio.wait_for(game_started.wait(), None)
    
    # At this point, game has started
    # We want to now block until we receive a game over event
    # As the game state update events will continue to be received while we are blocked here
    
    await asyncio.wait_for(game_over.wait(), None)
    # Once we get here, we have received a game over state
    # so now we are wrapping up folks

    for subscriber in subscribers:
        subscriber.unsubscribe()


if __name__ == "__main__":
    main()
```

There are other ways to run your bot and listen for events, 
but the subscription method is the easiest to onboard and setup, with the most code examples.