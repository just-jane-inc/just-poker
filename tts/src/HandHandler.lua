--[[
  At the start of a hand, the blinds and are moved and their values are decided. This work
  is simple but it's nice to encapsulate it here.
  
  The end of a hand doesn't really "require" anything, but for sanity we can encapsulate
  cleanup actions here, such as reforming the Dealer's deck, ensuring chips are in the
  right place on the table, and whatever else is condusive to starting subsequent hands.
]]

local BlindManager
local DeckManager
local TurnManager

function init(params)
  BlindManager  = params.Services.BlindManager
  DeckManager   = params.Services.DeckManager
  EventBus      = params.Services.EventBus

  EventBus.call("subscribe", {
    event_type        = "hand_started",
    callback_params   = {
      service = self,
      callback = "onHandStarted"
    }
  })
end

do -- Config
  HAND_START_DELAY  = 2
  HAND_END_DELAY    = 2
end

--[[
  
  When a hand starts, the buttons are configured.
  The big/small blinds are repositioned and may have a new cost.
  The dealer button is also repositioned.
  
  params: See EventBus.onHandStart
]]
function onHandStarted(params)
  local data = params.data
  local onFinish_params = params.onFinish_params

  

  --[[
    Configuration of the buttons is technically instantaneous, though
    the graphical transformation of their location takes some time.
    Since the server fires hand_start and round_start pretty much at the
    same time, we provide a reasonable delay for this transformation to 
    occur.
  ]]

  BlindManager.call("configureButtonsFromHandStart", data)
  broadcastToAll("Starting hand #"..data.id, "White")
  Wait.time(
    function() onFinish_params.service.call(onFinish_params.callback, onFinish_params.args) end,
    HAND_START_DELAY
  )
end