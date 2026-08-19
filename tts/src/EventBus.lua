--[[
  Various services in this mod rely on events to get the information needed
  to perform actions on the table. We need to be able to guarantee that 
  services dependent on events receive them, and that their actions are 
  completed before processing the next event.

  This service operates entirely independent of any other service. It is a 
  model of events that the poker server transmits to inform us of the state 
  of the game. External services that act on that information can subscribe
  to those events via the methods provided here.
]]


HAND_IN_PROGRESS         = false
ROUND_IN_PROGRESS        = false

HAND_DATA_TYPE = "hand_data"

ROUND_DATA_TYPE = "round_data"
ROUND_TYPES = {
  setup     = true,
  pre_flop  = true,
  flop      = true,
  turn      = true,
  river     = true,
  completed = true
}

PLAYER_ACTION_DATA_TYPE = "player_action_data"
PLAYER_INTENTS = {
  ante   = true,
  check  = true,
  call   = true,
  raise  = true,
  all_in = true,
  fold   = true
}
PLAYER_CHIP_DENOMINATIONS = {
  ["1000"] = true,
  ["500"]  = true,
  ["100"]  = true,
  ["50"]   = true,
  ["10"]   = true
}

function init(params)
  log("EventBus initialized")
end

--[[
  External scripts can set callbacks to be invoked when a specific event happens.

  Callbacks are invoked in order of when they were added. A callback must finished before
  the next one is invoked.

  Callbacks receive a table with relevant information to the event, and parameters for 
  notifying this script that they are finished.
]]
do

-- This mutex is for this script to know if it is currently invoking external scripts.
local is_currently_iterating_callbacks = false
function isNotCurrentlyIteratingCallbacks() return not is_currently_iterating_callbacks end

-- The flow for external callbacks is dictated by this mutex.
local awaiting_external_script_action = false
function lockExternalScriptMutex() log("EventBus mutex locked") awaiting_external_script_action = true end
function unlockExternalScriptMutex()  log("EventBus mutex unlocked") awaiting_external_script_action = false end
function isExternalScriptMutexUnlocked() return not awaiting_external_script_action end

--[[
  When a subscriber is notified of an event, they must handle that event 
  and inform the EventBus when they have finished.
]]
local onFinishExternalScript_params = {
  service  = self,
  callback = "unlockExternalScriptMutex",
  args     = nil
}

--[[
  External callers provide onFinish_params that describe what to do when
  finished.
  This is unreadable. Putting it in a function to hide away.
]]
function processOnFinishParams(onFinish_params)
  onFinish_params.service.call(onFinish_params.callback, onFinish_params.args)
end

-- -- hand_id, big/small_blind_cost, big/small_blind_position, button_position
-- local ON_HAND_START_CALLBACKS        = {}
-- -- hand_id
-- local ON_HAND_END_CALLBACKS          = {}
-- -- round_type
-- local ON_ROUND_START_CALLBACKS       = {}
-- -- round_type
-- local ON_ROUND_END_CALLBACKS         = {}
-- -- player_id, intent, chips
-- local ON_PLAYER_TURN_START_CALLBACKS = {}
-- -- player_id
-- local ON_PLAYER_TURN_END_CALLBACKS   = {}

--[[
  Went ahead and defined all expected event types that the server sends.

  The data field for these events is defined in relevant sections,
  but for more information you can review the Swagger:
  - See https://api.bahms.org/swagger/poker
]]
local EVENT_SUBSCRIPTIONS = {
  ["listener_created"]  = {},
  ["welcome"]           = {},
  ["starting_game"]     = {},
  ["game_state_update"] = {},
  ["hand_started"]      = {},
  ["round_start"]       = {},
  ["player_action"]     = {},
  ["payout"]            = {},
  ["game_over"]         = {}
}

--[[
  Subscribes a callback to an event defined in EventBus.EVENT_SUBSCRIPTIONS.
  
  Subscribers are called in the order they subscribed. The callback they 
  register receives the data from the payload carrying the event.

  params = {
    event_type = string,

    callback_params = {
      service = <Scripting Zone>,
      callback = string
    }
  }
]]
function subscribe(params)
  local event_type = params.event_type
  local callback_params = params.callback_params

  if not event_type then
    log({"No event_type provided", "EventBus.createSubscription"}, "[ERROR]", false, true)
    return
  end
  if not EVENT_SUBSCRIPTIONS[event_type] then
    log({"No publisher of " .. event_type .. " found"}, "EventBus.createSubscription", "[WARNING]", false, true)
  end
  if not callback_params then
    log({"Cannot subscribe to "..event_type.." without providing a callback"}, "EventBus.createSubscription", "[ERROR]", false, true)
    return
  end

  if not EVENT_SUBSCRIPTIONS[event_type] then
    EVENT_SUBSCRIPTIONS[event_type] = {}
  end

  table.insert(EVENT_SUBSCRIPTIONS[event_type], callback_params)
  log(callback_params.service, "Subscription to "..event_type.." created", "[INFO]", false, true)
end

--[[
  Publishes an event to all current subscribers.

  params = {
    event_type = string,
    data       = any,

    -- Optional - Only necessary for asynchronous operations.
    onFinish_params = {
      service  = <Scripting Zone>,
      callback = string,
      args     = any
    }
  }
]]
function publish(params)
  if not params.event_type then
    log("Missing params.event_type", "EventBus.publish", "[WARNING]", false, true)
  end
  if not EVENT_SUBSCRIPTIONS[params.event_type] then
    log("Event ["..params.event_type.."] has no subscribers", "EventBus.publish", "[WARNING]", false, true)
  end
  log("[EventBus.publish]: Publishing " .. params.event_type)
  publishToSubscribers(EVENT_SUBSCRIPTIONS[params.event_type], params.data)

  --[[
    If the caller provided a callback it will be invoked once all subscribers
    have been informed and processed the event.
  ]]
  if params.onFinish_params then
    local onFinish_params = params.onFinish_params
    Wait.condition(
      function() onFinish_params.service.call(onFinish_params.callback, onFinish_params.args) end,
      isNotCurrentlyIteratingCallbacks
    )
  end
end

--[[
  

  callback_params_array = [
    {
      service  = <Scripting Zone>,
      callback = string
    },
    ...
  ],

  data = any (the data being published, if any)
]]
function publishToSubscribers(callback_params_array, data)
  -- Don't want to block if there's nothing to do.
  if not callback_params_array or #callback_params_array == 0 then
    return
  end
  -- Locking this mutex informs this script that it cannot proceed
  -- the game until external work is finished.
  is_currently_iterating_callbacks = true

  local index = 1 -- Acts as an ID for the current callback being executed
  for i, callback_params in ipairs(callback_params_array) do
    local id = i
    local callback_params = callback_params
    Wait.condition(
      function()
        lockExternalScriptMutex()
        log(callback_params, "Publishing to: ", "", false, true)
        -- When this callback finishes, it will unlock the mutex, incrementing the index and allowing the next callback to be invoked
        callback_params.service.call(callback_params.callback, {data=data, onFinish_params=onFinishExternalScript_params})
        -- Once all callbacks have returned, the associated mutex isNotCurrentlyIteratingCallbacks is unlocked.
        Wait.condition(function()
          index = index + 1
          if index > #callback_params_array then
            is_currently_iterating_callbacks = false
          end
        end, isExternalScriptMutexUnlocked)
      end,
      function() return id == index and isExternalScriptMutexUnlocked() end
    )
  end
end

end -- End callbacks block

--[[
  Explicit coordination between scripts regarding actions that may take an indeterminate amount
  of time require blocking of the flow of execution to avoid race conditions.
  This paradigm is present in most other scripts in this mod, first defined in GameManager and
  replicated in other scripts as the need arose.

  There are two components to this:
  1. onFinish_params for the encapsulating script:
      {
        service  = self,
        callback = "unlockExternalScriptMutex"
      }
  2. The ability to process others' onFinish_params that follow the same paradigm.
]]
do

end

--[[
params = {
  data = {
    id = int, -- Hand ID
    
    -- Big blind settings
    big_blind_cost     = int,
    big_blind_position = int,

    -- Small blind settings
    small_blind_cost     = int,
    small_blind_position = int,

    -- Dealer button position
    button_position = int
  },

  onFinish_params = {
    service   = <Scripting Zone>,
    callback  = string (function name in service's script),
    args      = table (arguments to pass to callback)
  }
}
--]]
function startHand(params)
  log("Starting hand [TurnManager.startHand]")
  if not dataIsValid(HAND_DATA_TYPE, params.data) then
    log(params.data, "Could not validate hand_started data. ", "[ERROR]", false, true)
    broadcastToAll("Error: Failed to parse hand_start event", "Red")
    return
  end

  onHandStart(params.data)
  Wait.condition(
    function()
      Wait.time(
        function()
          processOnFinishParams(params.onFinish_params)
        end, HAND_START_DELAY
      )
      -- BlindManager.call("configureButtonsFromHandStart", params.data)
    end, isNotCurrentlyIteratingCallbacks
  )
end

--[[
params = {
  data = {
    id = int, -- Hand ID
    
    -- Big blind settings
    big_blind_cost     = int,
    big_blind_position = int,

    -- Small blind settings
    small_blind_cost     = int,
    small_blind_position = int,

    -- Dealer button position
    button_position = int
  },

  onFinish_params = {
    service   = <Scripting Zone>,
    callback  = string (function name in service's script),
    args      = table (arguments to pass to callback)
  }
}
--]]
function endHand(params)
  log("Ending hand TurnManager.endHand")

  onHandEnd(params.data)
  Wait.condition(
    function()
      broadcastToAll("End of hand #" .. params.data.id, "White")
      Wait.time(
        function()
          processOnFinishParams(params.onFinish_params)
        end, HAND_END_DELAY
      )
      -- DeckManager.gatherCards
    end, isNotCurrentlyIteratingCallbacks
  )
end

--[[
  This function is invoked from the GameManager upon receiving a "round_start" event.

  When a round starts--depending on the round, certain actions should take place:
  - pre_flop: Deal a hand to all active players,
  - flop|turn|river: Reveal a subset of the street

  The GameManager provides 2 fields in the "params" parameter:
  params = {
    round_type = string (whatever round has started)
  
    onFinish_params = {
      service   = <Scripting Zone>,
      callback  = string (function name in service's script),
      args      = table (arguments to pass to callback)
    }
  }
--]]
function startRound(params)
  if not dataIsValid("round_data", params.data) then
    log(params.data, "Could not validate round_start data. ", "[ERROR]", false, true)
    broadcastToAll("Error: Failed to parse round_start event.", "Red")
    return
  end
  -- External scripts must register a receiver to get information about the current round when it starts.
  onRoundStart(params.data.round_type)

  -- Once all callbacks are processed, this script announces to the table the current round.
  Wait.condition(
    function()
      broadcastToAll("Starting [" .. params.data.round_type .. "] round", "White")

      -- Todo: ensure these scripts have registered their receivers.
      -- if params.data.round_type ~= "setup" and params.data.round_type ~= "pre_flop" and params.data.round_type ~= "completed" then
      --   PotManager.call("gatherBets")
      --   PlayerManager.call("resetAllPlayerBets")
      -- end
      Wait.time(
        function()
          processOnFinishParams(params.onFinish_params)
        end, isNotCurrentlyIteratingCallbacks
      )
    end,
    isNotCurrentlyIteratingCallbacks
  )
end

--[[
  This function is invoked from the GameManager upon receiving a "player_action" event.
  TODO: When implementing human players, this will also need to be invoked upon the relevant notice.
  
  When a player's turn starts, the actions needed are
  1. Determine whose turn it is
  3. Lock their assets
  3. Locate their chips on the table
  4. Figure out how much they must bet to stay in, and what actions are available to them.
  5. For human players, render UI elements for the player to inform them of their state and give them the ability to end their turn.
  6. For human players, start listening for events related to impulses they make on the board.
  
  This information will be determined elsewhere, but a hook here allow ncessary services to run before they take any action.

  params = {
    -- Right now, only bots are able to play. Bots turns are completed before we receive them, so all the information necessary
    -- to play out their turn is available. The PlayerManager will handle this.
    data = {
      player_id = string,
      intent    = string,
      chips     = {[denomination] = int}
    },

    onFinish_params = {
      service   = <Scripting Zone>,
      callback  = string (function name in service's script),
      args      = table (arguments to pass to callback)
    }
  }
]]
-- function startPlayerTurn(params)
--   if not dataIsValid("player_params", params.data) then
--     log(params.data, "Could not validate player_action data. ", "[ERROR]", false, true)
--     broadcastToAll("Error: Failed to parse player action.", "Red")
--     return
--   end
--   -- External scripts must register a receiver to get information about the current round when it starts.
--   onPlayerTurnStart(params.data.player_id)
--   local player = PlayerManager.call("getUserById", params.data.player_id)
--   broadcastToAll(player.display_name.."'s turn", player.color)
  
--   Wait.condition(
--     function()
--       Wait.time(
--         function()
--           if (player.user_type == "bot") then
--             -- Bot actions are already performed. The PlayerManager needs to animate this for presentation.
--             -- To keep the turn from ending until this is accomplished, I'm locking the mutex again and letting
--             -- the PlayerManager release it when they have finished.
--             lockExternalScriptMutex()
--             PlayerManager.call("animatePlayerAction", {
--               data = params.data,
--               onFinish_params = TurnManager_onFinish_params
--             })

--             -- This condition is met when the PlayerManager invokes TurnManager_onFinish_params, which should
--             -- occur only when the player has successfully ended their turn.
--             Wait.condition(
--               function()
--                 onPlayerTurnEnd(params.data.player_id)
--                 Wait.condition(
--                   function()
--                     -- Todo: Map player states to a better worded broadcast.
--                     broadcastToAll(player.display_name .. " " .. player.state, player.color)
                    
--                     -- Finally we can return control to the GameManager.
--                     Wait.time( function() processOnFinishParams(params.onFinish_params) end, TURN_END_DELAY )
--                   end,
--                   isNotCurrentlyIteratingCallbacks
--                 )
--               end,
--               isExternalScriptMutexUnlocked
--             )
--           else
--             broadcastToAll("Error: Unsupported user type: ["..player.user_type.."]", "Red")
--             log("TurnManager.startPlayerTurn", "Unsupported user type for player ["..player.user_type.."]", "[ERROR]", false, true)
--             return
--           end
--         end, TURN_START_DELAY)
--     end,
--     isNotCurrentlyIteratingCallbacks
--   )
-- end

-- function startPreFlopRound()
--   broadcastToAll("Pre-flop", "White")
--   if DEALER_IS_AUTOMATED then
--     -- This call uses asynchronous coroutines to deal all the cards in 
--     -- a manner that looks natural. We need to block informing the caller
--     -- that we're done until this call has finished.
--     DeckManager.call("dealHand", TurnManager_onFinish_params)
--   end
-- end

-- function startFlopRound()
--   broadcastToAll("Flop", "White")
--   if DEALER_IS_AUTOMATED then
--     DeckManager.call("dealFlop", TurnManager_onFinish_params)
--   end
-- end

-- function startTurnRound()
--   broadcastToAll("Turn", "White")
--   if DEALER_IS_AUTOMATED then
--     DeckManager.call("dealTurn", TurnManager_onFinish_params)
--   end
-- end

-- function startRiverRound()
--   broadcastToAll("River", "White")
--   if DEALER_IS_AUTOMATED then
--     DeckManager.call("dealRiver", TurnManager_onFinish_params)
--   end
-- end

-- function startCompletedRound()
--   broadcastToAll("Hand over", "White")
--   if DEALER_IS_AUTOMATED then
--     DeckManager.call("gatherCards", TurnManager_onFinish_params)
--   end
-- end

-- When a player ends their turn we need to re-evaluate their chips
-- to ensure the ones in their stack are counted as their stack and
-- the ones in their bet are counted as their bet.
--[[
  This function is called when a player (directly, if it was a bot. See TurnManager.onPlayerTryEndTurn for human behavior)
  has ended their turn.

  For bots, this function is invoked from PlayerManager.animatePlayerAction.
]]
-- function endTurn(player_id)
--   local player = PlayerManager.call("getPlayerById", player_id)
--   if not player then
--     broadcastToAll("Error: Player with ID ["..player_id.."] does not exist.", "Red")
--     log(player_id, "TurnManager.endTurn", "[ERROR]", false, true)
--     return
--   end
-- -- if CurrentPosition == "Dealer" then
-- --     -- If it was the dealer's turn, the first player to act will be whoever has the small blind
-- --     local next_position = BlindManager.getVar("SmallBlindPosition")
-- --     local player_at_position = PlayerManager.call("getPlayerByPosition", next_position)
-- --     local player_with_first_turn_of_round = PlayerManager.call("getNextValidPlayer", player_at_position)
-- --     startTurnRound(player_with_first_turn_of_round)
-- --   end
--   log("Ending turn for " .. player.display_name)
--   ChipManager.call("mapPlayerChips", player)
-- end

--[[
  Determines whether the passed data is valid for the described data type.

  Valid "data_type":
    - "hand_data"
    - "round_data"

  See function body for expected data makeup. 
--]]
function dataIsValid(data_type, data)
  local data_is_valid = true
  local errors = {missing_fields={}, invalid_types={}}
  function addMissingField(field)
    data_is_valid = false
    table.insert(errors.missing_fields, field)
  end
  function addInvalidType(key, received, expected)
    data_is_valid = false
    table.insert(errors.invalid_types, {[key] = "Expected " .. expected .. ", found " .. received .. " instead."})
  end

  if     HAND_DATA_TYPE  == data_type then
    -- id
    if not data.id then
      addMissingField("id")
    end
    if data.id and not type(data.id) == "number" then
      data_is_valid = false
      addInvalidType("id", type(data.id), "number")
    end
    -- big blind cost
    if not data.big_blind_cost then
      addMissingField("big_blind_cost")
    end
    if data.big_blind_cost and not type(data.big_blind_cost) == "number" then
      addInvalidType("big_blind_cost", type(data.big_blind_cost), "number")
    end
    -- big blind position
    if not data.big_blind_position then
      addMissingField("big_blind_position")
    end
    if data.big_blind_position and not type(data.big_blind_position) == "number" then
      addInvalidType("big_blind_position", type(data.big_blind_position), "number")
    end
    -- small_blind_cost
    if not data.small_blind_cost then
      addMissingField("small_blind_cost")
    end
    if data.small_blind_cost and not type(data.small_blind_cost) == "number" then
      addInvalidType("small_blind_cost", type(data.small_blind_cost), "number")
    end
    -- small blind position
    if not data.small_blind_position then
      addMissingField("small_blind_position")
    end
    if data.small_blind_position and not type(data.small_blind_position) == "number" then
      addInvalidType("small_blind_position", type(data.small_blind_position), "number")
    end
    -- dealer button
    if not data.button_position then
      addMissingField("button")
    end
    if data.button_position and not type(data.button_position) == "number" then
      addInvalidType("button", type(data.button), "number")
    end
  elseif ROUND_DATA_TYPE == data_type then
    if not data.round_type then
      addMissingField("round_type")
    end
    if data.round_type and not ROUND_TYPES[data.round_type] then
      local round_types = {}
      for round_type, _ in ROUND_TYPES do table.insert(round_types, round_type) end
      addInvalidType("round_type", data.round_type, table.concat(round_types, "|"))
    end
  elseif PLAYER_ACTION_DATA_TYPE == data_type then
    if not data.player_id then
      addMissingField("player_id")
    end
    if data.player_id and not type(data.player_id) == "string" then
      addInvalidType("player_id", data.player_id, "string")
    end
    if not data.intent then
      addMissingField("intent")
    end
    if data.intent and not PLAYER_INTENTS[data.intent] then
      local player_intents = {}
      for intent, _ in PLAYER_INTENTS do table.insert(player_intents, intent) end
      addInvalidType("intent", data.intent, table.concat(player_intents, "|"))
    end
    if data.chips and not type(data.chips) == "table" then
      local denominations = {}
      for denomination, _ in PLAYER_CHIP_DENOMINATIONS do table.insert(denominations, denomination) end
      for denomination, count in data.chips do
        if not PLAYER_CHIP_DENOMINATIONS[denomination] then
          addInvalidType("chips", denomination, "index of "..table.concat(denominations, "|"))
        elseif not type(count) == "number" then
          addInvalidType('chips["'..denomination..'"]', type(count), "number")
        end
      end
    end
  end

  if not data_is_valid then
    log(errors, "Errors validating hand_start: ", "", false, true)
    return false
  end
  return true
end

--[[

  The stuff below here is probably obsolete.

]]

--[[
  Setup function for staging a round based on state
--]]
-- function setTurnFromTable(table)
--   CurrentBet = table.current_round.bet
--   CURRENT_ROUND_TYPE = table.current_round.current_round_type
--   CurrentPosition = table.current_round.current_player_position
--   local players = PlayerManager.call("getPlayers")
--   for player in players do
--     HAND_ZONES_TO_COLOR_MAP[player.zones.hand_zone] = player.color
--   end
-- end

do -- Real player region

-- function onPlayerTryEndTurn(player)
--   if playerInValidState(player) then
--     log("yes, you MAY end ur turn :^)")
--   else
--     log("do  more stuff bozo")
--   end
-- end

-- function playerInValidState(instance)
--   local player = PlayerManager.call("getPlayerByColor", player.color)
--   if not player then return false end

--   return true
-- end

-- do -- Event Handlers region
--   function onObjectEnterZone(zone, enter_object)
--     if enter_object.type ~= "Card" then return end
--     if HAND_ZONES_TO_COLOR_MAP[zone] == nil then return end

--     local player = PlayerManager.call("getPlayerByColor", HAND_ZONES_TO_COLOR_MAP[zone])
--     for card in player.hole do
--       if card == leave_object then
--         PLAYER_ACTION_MAP[player].cards_in_hand = PLAYER_ACTION_MAP[player].cards_in_hand + 1
--       end
--     end
--   end

--   -- Todo: Player.Hole should be a associative map array that way I can just index instead of iterating
  
--   function onObjectLeaveZone(zone, leave_object)
--     if leave_object.type ~= "Card" then return end
--     if HAND_ZONES_TO_COLOR_MAP[zone] == nil then return end
    
--     local player = PlayerManager.call("getPlayerByColor", HAND_ZONES_TO_COLOR_MAP[zone])
--     for card in player.hole do
--       if card == leave_object then
--         PLAYER_ACTION_MAP[player].cards_in_hand = PLAYER_ACTION_MAP[player].cards_in_hand - 1
--       end
--     end
--   end
-- end

end -- EndRegion of Real player