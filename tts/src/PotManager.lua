local ChipManager
local PotZone
local PotStackZone
local onTurnEnd_params

local LARGEST_BET_THIS_ROUND = 0
local POT_STACK = {
  ["1000"] = false,
  ["500"]  = false,
  ["100"]  = false,
  ["50"]   = false,
  ["10"]   = false
}

PLAYER_BETTED_CHIP_MAP = {
  --[[
    Makeup of this table:
    {
      ["user_id"] = {
        total_value_in_pot = int,
        bet                = int,
        
        chips = {
          ["denomination"] = {
            <Chip instance> = 1,
          },
          ...
        }
      }
      ...
    }
  ]]
  -- total_value_in_pot: int -- The total value of all chips in the pot
  -- bet: int -- The total value of chips being recognized as a bet
  -- <Chip instance>: int (1 if counted towards bet)
}

function init(params)
  ChipManager   = params.Services.ChipManager
  PlayerManager = params.Services.PlayerManager
  PotZone       = params.Zones.pot_zone
  PotStackZone  = params.Zones.pot_stack_zone

  onTurnEnd_params = {
    service  = self,
    callback = "updatePotStack",
    args     = nil
  }

  onRoundEnd_params = {
    service  = self,
    callback = "gatherBets",
    args     = nil
  }

  log("PotManager initialized")
end

do -- Config

GATHER_BETS_DELAY = 1.2

end


do -- Mutexes

local awaiting_payout = false
function lockPayoutMutex() awaiting_payout = true end
function unlockPayoutMutex() awaiting_payout = false end
function isPayoutMutexUnlocked() return not awaiting_payout end

local associative_mutex = {}
function lockAssociativeMutex(index) associative_mutex[index] = true end
function unlockAssociativeMutex(index) associative_mutex[index] = false end
function isAssociativeMutexUnlocked(index) return not associative_mutex[index] end

end

function setPotFromState(table)
  ChipManager.call("initializePot", table.pot)
  CurrentBet = table.current_round.bet
end



-- Determines if an object is a poker chip that we manage.
function isChip(object)
  return string.find(object.type, "Chip") and object.memo
end

function isBettedChip(object)
  return object ~= nil and object.memo ~= "Pot" and isChip(object) and string.find(object.memo, "(Bet)")
end

--[[
  player = <Player>,
--]]
function getBettedChipsForPlayer(player)
  return PLAYER_BETTED_CHIP_MAP[player.user_id].chips
end

-- From all the chips on the table, get references to those that are part of the pot.
function updatePotStack()
  POT_STACK = {
    ["1000"] = {},
    ["500"]  = {},
    ["100"]  = {},
    ["50"]   = {},
    ["10"]   = {}
  }
  for object in getObjects() do
    if isChip(object) and object.memo == "Pot" then
      table.insert(POT_STACK[tostring(object.getValue())], object)
    end
  end
end

--[[
  
]]
function gatherBets(onFinish_params)
  local offset = ChipManager.getVar("CHIP_DENOMINATION_RELATIVE_POSITION")
  for object in getObjects() do
    if isBettedChip(object) then
      local denomination = tostring(object.getValue())
      object.setName("Pot")
      object.memo = ("Pot")
      object.setPositionSmooth(PotStackZone.positionToWorld(offset[denomination]), false, false)
    end
  end

  Wait.time(
    function()
      updatePotStack()
      Wait.time(
        function() onFinish_params.service.call(onFinish_params.callback) end,
        GATHER_BETS_DELAY
      )
    end,
    1.5
  )
end

--[[
  Groups the pot in the middle and updates the references to its stacks.
]]
function updatePotStack()
  local pot_stack = {
    ["1000"] = {},
    ["500"]  = {},
    ["100"]  = {},
    ["50"]   = {},
    ["10"]   = {}
  }
  for object in getObjects() do
    if isChip(object) and object.memo == "Pot" then
      local denomination = tostring(object.getValue())
      table.insert(pot_stack[denomination], object)
    end
  end

  local offset = ChipManager.getVar("CHIP_DENOMINATION_RELATIVE_POSITION")
  for denomination, stacks in pairs(pot_stack) do
    if #stacks == 0 then pot_stack[denomination] = false
    elseif #stacks == 1 then
      pot_stack[denomination] = stacks[1]
      pot_stack[denomination].setPosition(PotStackZone.positionToWorld(offset[denomination]))
    elseif #stacks >= 2 then
      pot_stack[denomination] = group(stacks)[1]
      pot_stack[denomination].setPosition(PotStackZone.positionToWorld(offset[denomination]))
    end
  end

  POT_STACK = pot_stack
end

--[[
  payoutChips_params = {
    player_id = string,
    chips     = {
      "1000" = int,
      "500"  = int,
      "100"  = int,
      "50"   = int,
      "10"   = int,
    },

    onFinish_params = {
      service = <Scripting Zone>
      callback = string,
      args = any
    }
  }  
]]
function animatePayout(payoutChips_params)
  local player = PlayerManager.call("getPlayerById", payoutChips_params.player_id)
  if not player then
    log("Could not find player with this ID: " .. payoutChips_params.player_id, "PotManager.animatePayout", "[ERROR]", false, true)
    return
  end

  local chips = payoutChips_params.chips

  local promises = {}
  local denominations = {}
  local counts = {}
  for denomination, count in pairs(chips) do
    table.insert(denominations, denomination)
    table.insert(counts, count)
    table.insert(promises, false)
  end

  local onFinish_params = payoutChips_params.onFinish_params
  
  if not chips or #denominations == 0 then
    -- No chips to move.
    log("No chips paid out to "..player.display_name, "PotManager.animatePayout", "[WARNING]", false, true)
    onFinish_params.service.call(onFinish_params.callback, onFinish_params.args)
    return
  end

  lockPayoutMutex()
  updatePotStack()

  do -- Internal mutexes

    -- If not all denominations have been allocated then this returns false
    function _promisesFulfilled()
      for promise_fulfilled in promises do if not promise_fulfilled then return false end end
      return true
    end
  
  end

  local player_stack_zone = player.zones.stack_zone
  local chip_offsets      = ChipManager.getVar("CHIP_DENOMINATION_RELATIVE_POSITION")

  function _getChipForDenomination(denomination)
    if not POT_STACK[denomination] then
      log("No $"..denomination.." chips in pot", "PotManager.animatePayout", "[ERROR]", false, true)
      return
    end

    if POT_STACK[denomination] then
      return POT_STACK[denomination]
    end
    log("No valid reference to $"..denomination.." chips in pot", "PotManager.animatePayout", "[ERROR]", false, true)
  end

  function _getStackDestinationForDenomination(denomination)
    PlayerManager.call("updatePlayerStackAndBet", player)
    if player.stack[denomination] then
      -- Player has chips of this denomination. We'll use them as a target
      for chip_reference in player.stack[denomination] do
        if chip_reference then
          return chip_reference
        end
      end
    end
    return player_stack_zone.positionToWorld(chip_offsets[denomination])
  end

  function _transferChipsForDenomination(index)
    local denomination = denominations[index]
    local count        = counts[index]

    if count == 0 then
      promises[index] = true
      return
    end

    local chip_reference = _getChipForDenomination(denomination)
    local destination    = _getStackDestinationForDenomination(denomination)

    if not chip_reference then
      log("No $"..denomination.." chip to move", "PotManager.animatePayout", "[ERROR]", false, true)
      return
    end

    local new_name = "- "..player.display_name
    local new_memo = player.user_id

    local transferChips_params = {
      chips_to_transfer = chip_reference,
      destination       = destination,
      count             = count,
      new_name          = new_name,
      new_memo          = new_memo,
      onFinish_params   = {
        service  = self,
        callback = "unlockAssociativeMutex",
        args     = denomination
      }
    }

    lockAssociativeMutex(denomination)
    local num_of_chips_moved = ChipManager.call("transferChips", transferChips_params)
    counts[index] = count - num_of_chips_moved

    Wait.condition(
      function()
        updatePotStack()
        _transferChipsForDenomination(index)
      end,
      function() return isAssociativeMutexUnlocked(denomination) end
    )
  end

  for i = 1, #promises do
    _transferChipsForDenomination(i)
  end

  Wait.condition(
    function() onFinish_params.service.call(onFinish_params.callback, onFinish_params.args) end,
    _promisesFulfilled
  )

end

--[[
  Pay out chips from the pot to winning players.

  params = {
    data = array[
      {
        player_id = string,
        chips: {
          "1000" = int,
          "500"  = int,
          "100"  = int,
          "50"   = int,
          "10"   = int
        }
      },
      ...
    ],

    onFinish_params = {
      service  = <Scripting Zone>,
      callback = string,
      args     = table|nil
    }
  }
]]
function processPayoutOld(params)
  -- TODO: Verify that the state of the pot in TTS doesn't differ from the state of the pot
  -- on the server. Players are able to swap chips between their bet and stack during their
  -- turn and we need to ensure these line up if I'm to reuse the same chips.

  -- This variable is defined below in the outer scope.
  -- This is done so object spawn events can be recognized and associated with this operation
  -- and the spawned chip can be added to the references used here.
  pot_chips = {
    ["1000"] = {},
    ["500"]  = {},
    ["100"]  = {},
    ["50"]   = {},
    ["10"]   = {}
  }
  
  -- don't care where the chips are. if they have the "Pot" memo they belong to the pot.
  for object in getObjects() do
    if object.memo and object.memo == "Pot" then
      pot_chips[tostring(object.getValue())][object] = true
    end
  end

  -- Blocks the next player from getting paid out until this player has finished being
  -- payed out.
  local payout_occuring = false
  function noPayoutOccuring() return not payout_occuring end

  -- Everything happening here has to work asynchronously.
  -- This array let's me use a JS-like Promise.all(...) approach.
  local finished_payouts = {}
  for _ in params.data do table.insert(finished_payouts, false) end

  -- A payout to a player involves allocating the correct amount and type of chips
  -- from the pot, moving them to the player's stack (or where I think it should be),
  -- and changing the metadata on the chip to assign ownership to that player.
  function payoutChipsToPlayer(payout_data, finished_payouts_index)
    payout_occuring = true
    local player = PlayerManager.call("getPlayerById", payout_data.player_id)

    -- I use indices to check where in the payout we are so we can release the mutex
    -- after the last chip of the last denomination is delivered. Gotta unpack the payout
    -- map first.
    local denominations = {}
    local counts        = {}
    -- Like the Promise.all approach for tracking the progress of all player payouts.
    -- Different denominations can be paid out at the same time, all that matters is
    -- that we know when they're all done.
    local finished_denominations = {}
    for denomination, count in pairs(payout_data.chips) do
      table.insert(denominations, denomination)
      table.insert(counts, count)
      table.insert(finished_denominations, false)
    end

    -- Because the handle for a stack can be lost at any time, getting the first "non-nil"
    -- value in this array is the most reliable way to ensure i don't get a random
    -- null-reference error.
    -- See payout_chips, onObjectSpawn, and onObjectDestroy defined at the end of this file.
    function getNextChip(denomination)
      for chips, valid_reference in pairs(pot_chips[denomination]) do
        if valid_reference then
          return chips
        end
      end
    end

    for i = 1, #denominations do
      -- Below are quasi-recursive co-routines that need a static reference to the current index
      -- in order to know which denomination and count they own.
      local index = i
      local denomination = denominations[index]
      local count = counts[index]

      -- This reference may update up to 3 times throughout execution of payoutChipsToPlayer per denomination
      -- In order of: Vector3, <Chip>, <ChipStack>
      local destination = ChipManager.call("getPositionOfChipsForPlayer", {player_id = player.user_id, denomination = denomination})
      
      
      -- In TTS there's several permutations of how to move one object to another location.
      -- 1. Is the object being moved from a stack, or a single?
      -- 2. Is the destination a position, or an existing object?
      -- We need to maintain a reference to "where" the chip was moved, whether thats the original
      -- chip or a new stack.
      function moveChipToDestination(chip_to_move, onFinish_callback)
        -- The destination is a Vector3. 
        if not destination.type then
          if chip_to_move.getQuantity() == -1 then
            -- Single chips are moved with setPosition*
            chip_to_move.setPositionSmooth(destination, false, false)
            Wait.time(onFinish_callback, 0.17) -- This will call approximately after the chip finishes travelling visually
            ChipManager.call("assignChipsToPlayer", {player_id = player.user_id, chips = chip_to_move})
            PlayerManager.call("addChipToStack", chip_to_move)
          else
            -- A chip from a stack is moved using takeObject
            return chip_to_move.takeObject({
              position = destination,
              top = true,
              smooth = false,
              callback_function = function(chip)
                ChipManager.call("assignChipsToPlayer", {player_id = player.user_id, chips = chip})
                PlayerManager.call("addChipToStack", chip)
                destination = chip
                Wait.time(onFinish_callback, 0.17) -- This will call approximately after the chip finishes travelling visually
              end
            })
          end
        else -- Destination is a chip
          local destination_quantity = destination.getQuantity() -- If this is a single chip, then a new stack is formed and needs to be tracked
          if chip_to_move.getQuantity() == -1 then
            ChipManager.call("assignChipsToPlayer", {player_id = player.user_id, chips = chip_to_move})
            if destination_quantity == -1 then
              PlayerManager.call("removeChipFromStack", destination) -- player.stack's reference to the destination chip becomes invalid once a stack is formed
              destination = destination.putObject(chip_to_move)
              PlayerManager.call("addChipToStack", destination) -- A reference to the new stack is added in its place.
            else
              destination.putObject(chip_to_move)
            end
            onFinish_callback()
          else
            chip_to_move.takeObject({
              top = true,
              smooth = false,
              callback_function = function(chip)
                ChipManager.call("assignChipsToPlayer", {player_id = player.user_id, chips = chip})
                if destination_quantity == -1 then
                  PlayerManager.call("removeChipFromStack", destination)
                  destination = destination.putObject(chip)
                  PlayerManager.call("addChipToStack", destination)
                end
                onFinish_callback()
              end
            })
          end
        end
      end

      -- Each chip for a denomination must be moved one-at-a-time. Explanation is noted with pot_chips
      local count_index = 1
      for c = 1, count do
        -- local assignment gives the conditional a static reference to ci at this point in time
        local ci = c

        -- Each chip transfer gets its own coroutine that waits until it is their turn to find a
        -- reference for the chip that is being moved.
        Wait.condition(
          function()

            local chip_reference = getNextChip(denomination)
            if chip_reference == nil then
              log("PotManager.processPayout.payoutChipsToPlayer", "Error: Failed to get chip_reference", "[ERROR]", false, true)
              return
            end

            if count_index < count then
              moveChipToDestination(chip_reference, function() count_index = count_index + 1 end)
            else
              -- Once count_index equals count, all chips for this denomination have been moved.
              moveChipToDestination(chip_reference, function() finished_denominations[index] = true end)
            end
          end,
          function() return count_index == ci end
        )
      end
    end
  end


  for index, payout_data in ipairs(params.data) do
    payoutChipsToPlayer(payout_data, index)
  end

  Wait.condition(
    function()
      params.onFinish_params.service.call(params.onFinish_params.callback, params.onFinish_params.args)
    end,
    function()
      -- If all indices in finished_payouts are true, all payouts have finished. Duh :^)
      for is_complete in finished_payouts do
        if not is_complete then return false end
      end
      return true
    end
  )
end

--[[
  This is a reference to the chips being processed during a payout. If its nil, there's not a payout
  happening. If it isn't nil, a payout is happening.

  Keeping this reference allows me to track the "creation" and "destruction" of chips during payout
  (such as when all but one chip is taken out of a stack, and when a chip is added to another player's
  stack.
]]
local pot_chips

function onObjectSpawn(object)
  if pot_chips and object.memo and object.memo == "Pot" then
    pot_chips[tostring(object.getValue())][object] = true
  end
end
function onObjectDestroy(object)
  if pot_chips and pot_chips[object] and object.memo and object.memo == "Pot" then
    pot_chips[tostring(object.getValue())][object] = nil
  end
end

--[[
  This handles tracking the value of all the chips a specific player has
  put into the pot.
]]
function adjustCurrentBetValueForId(player_id)
  PLAYER_BETTED_CHIP_MAP[player_id].bet = 0
  for _, chips in pairs(PLAYER_BETTED_CHIP_MAP[player_id].chips) do
    for chip, _ in pairs(chips) do
      local chip_value = chip.getValue() * math.abs(chip.getQuantity())
      PLAYER_BETTED_CHIP_MAP[player_id].bet = PLAYER_BETTED_CHIP_MAP[player_id].bet + chip_value
    end
  end
end

function initializeChipsMapForPlayerId(player_id)
  if not PLAYER_BETTED_CHIP_MAP[player_id] then PLAYER_BETTED_CHIP_MAP[player_id] = {} else return end
  if not PLAYER_BETTED_CHIP_MAP[player_id].bet then PLAYER_BETTED_CHIP_MAP[player_id].bet = 0 end
  if not PLAYER_BETTED_CHIP_MAP[player_id].total_value_in_pot then PLAYER_BETTED_CHIP_MAP[player_id].total_value_in_pot = 0 end
  if not PLAYER_BETTED_CHIP_MAP[player_id].chips then PLAYER_BETTED_CHIP_MAP[player_id].chips = {["1000"] = {}, ["500"] = {}, ["100"] = {}, ["50"] = {}, ["10"] = {}} end
end

-- These handle chips going in/out of the zone.
function onObjectEnterScriptingZone(zone, object)
  if zone ~= PotZone then return end
  if not isChip(object) then return end
  if object.memo == "Pot" then return end
  if not PLAYER_BETTED_CHIP_MAP[object.memo] then return end
  PLAYER_BETTED_CHIP_MAP[object.memo].chips[tostring(object.getValue())][object] = 0
  -- log("Userid " .. object.memo .. " has added $" .. object.getValue()*math.abs(object.getQuantity()) .. " to their bet")
  adjustCurrentBetValueForId(object.memo)
end
function onObjectLeaveScriptingZone(zone, object)
  if zone ~= PotZone then return end
  if not isChip(object) then return end
  if object.memo == "Pot" then return end
  if not PLAYER_BETTED_CHIP_MAP[object.memo] then return end
  PLAYER_BETTED_CHIP_MAP[object.memo].chips[tostring(object.getValue())][object] = nil
  -- log("Userid " .. object.memo .. " has added -$" .. object.getValue()*math.abs(object.getQuantity()) .. " to their bet")
  adjustCurrentBetValueForId(object.memo)
end

-- These handle chips being added/removed from a stack
function onObjectEnterContainer(container, enter_object)
  if not PLAYER_BETTED_CHIP_MAP[container.memo] then return end
  if isBettedChip(container) or isBettedChip(enter_object) then
    if enter_object.memo == "Pot" then return end
    PLAYER_BETTED_CHIP_MAP[container.memo].chips[tostring(container.getValue())][enter_object] = nil
    adjustCurrentBetValueForId(container.memo)
  end
  -- log("Userid " .. enter_object.memo .. " has added $" .. enter_object.getValue()*math.abs(enter_object.getQuantity()) .. " to an existing stack now equaling $" .. container.getValue()*math.abs(container.getQuantity()))
end
function onObjectLeaveContainer(container, leave_object)
  if not PLAYER_BETTED_CHIP_MAP[container.memo] then return end
  if isBettedChip(container) or isBettedChip(leave_object) then
    if leave_object.memo == "Pot" then return end
    adjustCurrentBetValueForId(container.memo)
  end
-- log("Userid " .. leave_object.memo .. " has removed -$" .. leave_object.getValue()*math.abs(leave_object.getQuantity()) .. " from an existing stack now equaling $" .. container.getValue()*math.abs(container.getQuantity()))
end