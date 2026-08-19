
local BlindManager
local ChipManager
local DeckManager
local DealerDeckZone
local PlayerManager
local PotManager
local EventBus

-- Prefab deck layout for testing DeckManager.sortDeck
local PREFAB_DECK_LAYOUT = '[{"rank":50,"suit":99},{"rank":50,"suit":104},{"rank":57,"suit":99},{"rank":84,"suit":99},{"rank":74,"suit":115},{"rank":52,"suit":100},{"rank":53,"suit":104},{"rank":52,"suit":104},{"rank":57,"suit":100},{"rank":53,"suit":115},{"rank":81,"suit":115},{"rank":57,"suit":104},{"rank":75,"suit":99},{"rank":53,"suit":99},{"rank":55,"suit":99},{"rank":65,"suit":99},{"rank":75,"suit":100},{"rank":65,"suit":115},{"rank":84,"suit":115},{"rank":65,"suit":104},{"rank":51,"suit":99},{"rank":51,"suit":104},{"rank":75,"suit":104},{"rank":54,"suit":99},{"rank":55,"suit":115},{"rank":55,"suit":104},{"rank":56,"suit":100},{"rank":56,"suit":104},{"rank":54,"suit":115},{"rank":74,"suit":104},{"rank":81,"suit":104},{"rank":51,"suit":115},{"rank":57,"suit":115},{"rank":54,"suit":100},{"rank":56,"suit":99},{"rank":52,"suit":99},{"rank":75,"suit":115},{"rank":84,"suit":104},{"rank":52,"suit":115},{"rank":84,"suit":100},{"rank":55,"suit":100},{"rank":50,"suit":115},{"rank":81,"suit":99},{"rank":51,"suit":100},{"rank":74,"suit":100},{"rank":81,"suit":100},{"rank":65,"suit":100},{"rank":50,"suit":100},{"rank":53,"suit":100},{"rank":54,"suit":104},{"rank":74,"suit":99},{"rank":56,"suit":115}]'

--[[
    Ran into an issue where I couldn't access the PotManager by
    GUID in the onLoaded overload.
    Maybe this script runs before all objects are loaded?
    Just gonna hand off initialization to the Global script
    if I want access to other scripts.
]]
function init(params)
    BlindManager    = params.Services.BlindManager
    ChipManager     = params.Services.ChipManager
    EventBus        = params.Services.EventBus
    DeckManager     = params.Services.DeckManager
    PlayerManager   = params.Services.PlayerManager
    PotManager      = params.Services.PotManager

    do -- EventBus Subscriptions
    EventBus.call("subscribe", {
      event_type = "welcome",
      callback_params = {
        service  = self,
        callback = "onWelcome"
      }
    })
    EventBus.call("subscribe", {
      event_type = "starting_game",
      callback_params = {
        service  = self,
        callback = "onStartingGame"
      }
    })
    EventBus.call("subscribe", {
      event_type = "game_state_update",
      callback_params = {
        service  = self,
        callback = "onGameStateUpdate"
      }
    })
    end
    
    log("GameStateHandler initialized")
end

do -- Config
  GAME_STATE_LOADED_DELAY = 1
end

--[[
  Since many actions in Tabletop Simulator are asynchronous, and lua doesn't have an
  async/await paradigm (that i know how to implement), I came up with a paradigm that
  allows me to block the flow of execution based on a mutex from the caller.

  When external services (other scripts) are called, I pass them a table containing
  expected data and the following table. This table describes what the external script
  should invoke upon completion of its actions.
]]
local GameManager_onFinish_params = {
  service = self,
  callback = "unlockExternalScriptMutex",
  args = false
}

-- If the game is known to be in progress then there is already a game_state loaded,
-- no additional game_state should be parsed.
local GAME_IN_PROGRESS = false

-- If the game is over, events of this game_id should be ignored
local GAME_OVER = false


function onWelcome(params)
  log("Welcome event received") -- DEBUG

  if GAME_OVER then
    log("Game is over, ignoring game_state_update")
    return
  end

  -- If the state was provided by the welcome message, all players may not have joined yet.
  -- If so, the game hasn't started yet and we can back out and wait for the "game_start"
  -- event to get a more complete view of the table.
  if not params.data.started_at then
    local onFinish_params = params.onFinish_params
    log("Game not started yet. Deferring game_state load to game_start event.")
    log(params)
    onFinish_params.service.call(onFinish_params.callback)
    return
  end
  onGameStateReceived(params)
end

function onStartingGame(params)
  log("game_started message received") -- DEBUG
  onGameStateReceived(params)
end

function onGameStateUpdate(params)
  log("game_state_update event received") -- DEBUG
  onGameStateReceived(params)
end

--[[
  Game state can be set on two conditions:
    1. The "game_started" event is received (New game)
    2. GAME_IN_PROGRESS is false when a game_state_update event is received
]]
function onGameStateReceived(params)
  local data = params.data
  local onFinish_params = params.onFinish_params

  if GAME_IN_PROGRESS then
    -- We don't get events specifically for when a player is out
    for player_data in params.data.table.players do
      local player = PlayerManager.call("getPlayerById", player_data.user_id)
      player.state = player_data.state
    end
    onFinish_params.service.call(onFinish_params.callback, onFinish_params.args)
    return
  end

  if not data.table or type(data.table) ~= "table" then
    print("Error: data.table is not an isntance of Table")
    return
  end
  local table = data.table

  -- TODO: The game state should provide the deck layout if a game is in progress.
  -- TODO: The cards in play should be dealt from the deck same as in normal game playback

  --[[
    Initializes player details:
      - Position (and which seat they're in)
      - Hands -- TODO: this should be removed when the deck data is provided
      - Stacks
  ]]
  PlayerManager.call("setPlayers", table.players)

  --[[
    Initializes the buttons:
      - Set costs
      - Set positions
  ]]
  BlindManager.call("configureButtonsFromTable", table)

  --[[
    Initializes the street:
      - Burns cards and reveals the Flop/Turn/River
  ]]
  DeckManager.call("setStreet", table)

  -- Build pot
  -- we should have bets be in chip denominations for my sanity
  PotManager.call("setPotFromState", table)

  -- Set the turn to the correct player and continue with the game
  -- TurnManager.call("setTurnFromTable", table) -- Obselete: Server tracks turns.

  Wait.time(
    function()
        -- Logic is gated by the value of this boolean.
        -- Stops some event handlers from going crazy until a game is actually started
        GAME_IN_PROGRESS = true
        onFinish_params.service.call(onFinish_params.callback, onFinish_params.args)
      end,
    GAME_STATE_LOADED_DELAY
  )
end

-- --[[
--   Starting a hand signifies:
--   1. The buttons moved and may have changed in costs
--   2. A new deck order was provided by us or the server.
-- ]]
-- function onHandStartReceived(data)
--   lockExternalScriptMutex()
--   local params = {
--     data = data,
--     onFinish_params = GameManager_onFinish_params
--   }

--   EventBus.call("startHand", params)
-- end

-- function onRoundStartReceived(data)
--   local params = {
--     data = data,
--     onFinish_params = GameManager_onFinish_params
--   }

--   EventBus.call("startRound", params)
-- end

-- function onPlayerActionReceived(data)
--   local params = {
--     data = data,
--     onFinish_params = GameManager_onFinish_params
--   }

--   PlayerManager.call("animatePlayerAction", params)
-- end

-- function onRoundEndReceived(data)
-- end

-- function onPayoutReceived(data)
--   local params = {
--     data = data,
--     onFinish_params = GameManager_onFinish_params
--   }
--   PotManager.call("processPayout", params)
-- end