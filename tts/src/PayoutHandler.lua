-- --[[
--     After a hand is played and a winner is decided we must divvy the chips. There
--     can be one or more winners and the chips awarded come from the pot.


-- ]]

-- local ChipManager
-- local EventBus
-- local PotManager

-- function init(params)
--   ChipManager   = params.Services.ChipManager
--   EventBus      = params.Services.EventBus
--   PotManager    = params.Services.PotManager
--   PlayerManager = params.Services.PlayerManager

--   EventBus.call("subscribe", {
--       event_type = "payout",
--       callback_params = {
--           service  = self,
--           callback = "onPayout"
--       }
--   })

--   log("PayoutHandler initialized")
-- end

-- do -- Mutexes

-- local awaiting_external_script = false
-- function lockExternalScriptMutex() awaiting_external_script = true end
-- function unlockExternalScriptMutex() awaiting_external_script = false end
-- function isExternalScriptMutexUnlocked() return not awaiting_external_script end

-- PayoutHandler_onFinish_params = {
--   service  = self,
--   callback = "unlockExternalScriptMutex"
-- }

-- end

-- --[[
-- The payload for payout looks like:
-- [
--   {
--     "player_id": string,
--     "chips" {
--       "1000": int,
--       ...
--     }
--   },
--   ...
-- ]

-- This should match what's in the pot. We'll just pay out from the pot and delete
-- any leftover chips.
-- ]]
-- function onPayout(params)
--   local data = params.data
--   local onFinish_params = params.onFinish_params

--   -- Right now we don't have an event specific to when players 

--   if #data == 0 then
--     -- Just some stupid edge case i thought of
--     onFinish_params.service.call(onFinish_params.callback, onFinish_params.args)
--     return
--   end

--   local index = 1
--   for i, payout_data in ipairs(data) do
--     local id = i
--     local payoutChips_params = {
--       player_id       = payout_data.player_id,
--       chips           = payout_data.chips,
--       onFinish_params = PayoutHandler_onFinish_params
--     }

--     Wait.condition(
--       function()
--         lockExternalScriptMutex()
--         PotManager.call("animatePayout", payoutChips_params)
--         index = index + 1
--       end,
--       function() return isExternalScriptMutexUnlocked() and index == id end
--     )
--   end

--   Wait.condition(
--     function() onFinish_params.service.call(onFinish_params.callback, onFinish_params.args) end,
--     function() return isExternalScriptMutexUnlocked() and index == #data + 1 end
--   )
-- end