--[[ Lua code. See documentation: https://api.tabletopsimulator.com/ --]]

--[[ The onLoad event is called after the game save finishes loading. --]]
LOAD_TEST_GAME = false
START_TEST_GAME = false

CHIP_DENOMINATIONS = {["1000"] = 0, ["500"] = 0, ["100"] = 0, ["50"] = 0, ["10"] = 0}

local Objects
local Services
local Zones

function onLoad()
  print("=============================") -- Helps delimit restarts

  function createPlayerZoneTable(hand_zone_guid, stack_zone_guid, bet_zone_guid, blind_zone_guid)
    return {
      hand_zone  = getObjectFromGUID(hand_zone_guid),
      stack_zone = getObjectFromGUID(stack_zone_guid),
      bet_zone   = getObjectFromGUID(bet_zone_guid),
      blind_zone = getObjectFromGUID(blind_zone_guid)
    }
  end

  -- Group services and zones into a globally accessible format
  Services = {
    ServerClientConnection = getObjectFromGUID("1f6a51"),

    BlindManager     = getObjectFromGUID("07e1ee"),
    ChipManager      = getObjectFromGUID("c56c2d"),
    DeckManager      = getObjectFromGUID("f7f04c"),
    EventBus         = getObjectFromGUID("0f9a75"),
    PlayerManager    = getObjectFromGUID("03b336"),
    PotManager       = getObjectFromGUID("72d9ba"),

    GameStateHandler = getObjectFromGUID("ff149d"),
    HandHandler      = getObjectFromGUID("dc1416"),
    RoundHandler     = getObjectFromGUID("d97696"),
    TurnHandler      = getObjectFromGUID("1ec0b2"),
    PayoutHandler    = getObjectFromGUID("c1d0da")
  }

  Zones = {
    PlayerZones = {
      Red    = createPlayerZoneTable("0679f3", "1db08a", "cad943", "6c8e98"),
      Orange = createPlayerZoneTable("b1c996", "b4776d", "79ac5d", "eae4c3"),
      Yellow = createPlayerZoneTable("a16e18", "bc5bd3", "61e3f4", "2f7658"),
      Green  = createPlayerZoneTable("52b418", "123649", "7c3a59", "5de3a9"),
      Teal   = createPlayerZoneTable("5a773c", "7526cf", "eb2f1e", "363913"),
      Blue   = createPlayerZoneTable("c05d40", "6723f4", "ffbfd9", "47a939"),
      Purple = createPlayerZoneTable("23c35a", "c19173", "0fc032", "016b77"),
      Pink   = createPlayerZoneTable("ff2ae5", "92cb00", "06b225", "5d5794"),
    },

    dealer_deck_zone = getObjectFromGUID("22b504"),
    pot_zone         = getObjectFromGUID("cac241"),
    pot_stack_zone   = getObjectFromGUID("8a44ce")
  }

  Objects = {
    initial_deck  = getObjectFromGUID("9988fb"),
    dealer_button = getObjectFromGUID("f26272"),
    big_blind     = getObjectFromGUID("a560f3"),
    small_blind   = getObjectFromGUID("a3ca23")
  }

  local params = {Services = Services, Zones = Zones, Objects = Objects}

  -- Initialize services
  log("Initializing services...")

  for service in params.Services do
    service.call("init", params)
  end
  
  log("All services loaded!")

  do -- Whatever buttons i created to help me test things

  -- Create UI
  local create_deck = getObjectFromGUID("b0cdb9")
  create_deck.createButton({
    label=          "Create a deck with metadata",
    click_function= "createDeck",
    function_owner= Services.DeckManager,
    position=       {0, 0.5, 1},
    height=300, width=1000, font_size=100
  })

  local find_ace_of_spades = getObjectFromGUID("cebd10")
  find_ace_of_spades.createButton({
    label=          "Sort deck",
    click_function= "sortDeck",
    function_owner= Services.DeckManager,
    position=       {0, 0.5, 1},
    height=300, width=1000, font_size=100
  })

  local get_deck_order = getObjectFromGUID("71225f")
  get_deck_order.createButton({
    label=          "Get deck order",
    click_function= "getCardsInDeck",
    function_owner= Services.DeckManager,
    position=       {0, 0.5, 1},
    height=300, width=1000, font_size=100
  })

  end

  if START_TEST_GAME then
    Services.ServerClientConnection.call("createEventListener")
  end
end

function onUpdate() end