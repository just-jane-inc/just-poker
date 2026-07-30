# Endpoints
## /player/{id} GET
 - is_bot
 - name
 - player_id

## /game/{id} GET

|name|type|description|
|:-|:-:|:-|
|time_started|datetime|two|
|current_state|GameState|two|


## /game/{id}/history? (oneof query) GET

|name|type|description|
|:-|:-:|:-|
|states|GameState[]|two|

## Data Structures

### GameState

|name|type|description|
|:-|:-:|:-|
|hand|Hand||
|turn|Turn||
|big_blind|int||
|small_blind|int||
|street|Card[]||
|current_pot|int|why|
|last_action|Action?|{player_id, chips, action enum}|
|button_at|string|player_id|
|players|Players[]|array representing table order, position 0 is always current player turn|


### Hand

|name|type|description|
|:-|:-:|:-|
|started|datetime|hi|
|ended|datetime?|set if the round is over?|
|count|int||

### Turn

|name|type|description|
|:-|:-:|:-|
|started|datetime|hi|
|ended|datetime?|set if the round is over?|
|count|int||

### Player
when this is fetched for in progress game cards show as XX unless the requesting player is authenticated as the player

|name|type|description|
|:-|:-:|:-|
|id|string|bahms-poker-player id for player|
|display_name|string|display name|
|cards|Card[]|set if the game is over?|
|chips|Chip{}|mapping of int->int denomination->count|
|chip_total|int|dumb|
|current_state|PlayerStateEnum|YOUR_TURN, FOLDED, ALL_IN, READY, WON, OUT, UNSET|


### Card
Text/String (2-Character): Most programming APIs use a straightforward two-character
format ([Rank][Suit]). The rank is represented as A, 2, 3, 4, 5, 6, 7, 8, 9, T, J, Q, K,
and suits are represented by s (Spades), h (Hearts), d (Diamonds), and c (Clubs). 
For example, the Ten of Spades is Ts and the Ace of Hearts is Ah

- Ah
- XX (unknown)
[Ah, As, Kd,XX,XX]


### /game/{id}/state POST Action

|name|type|description|
|:-|:-:|:-|
|player_id|string||
|bet_amount|int||
|action|ActionEnum|check, fold, raise, call|

response can be 200 OK or 400 with error code:
- not your turn
- not enough chips

### /game/{id}/chip_exchange POST


### Invalid Moves / Error Handling
three strikes your out policy, error prone bots are removed from the game

### /game/{id}/player/{id}

|name|type|description|
|:-|:-:|:-|
|id|string|bahms-poker-player id for player|
|display_name|string|display name|
|cards|Card[]|set if the game is over?|
|chips|Chip{}|mapping of int->int denomination->count|
|chip_total|int|dumb|
|current_state|PlayerStateEnum|YOUR_TURN, FOLDED, ALL_IN, READY, WON, OUT, UNSET|


---

- start a new game
- /game POST {config} -> game_id
- /game/{id}/player POST {player} request to join, admin panel approves (game maker) or config to just auto allow for custom games
- /game/{id}/ws
- connections happen optionally
- we send start game signal
- bots start polling

- start a new game, includes set of players and chips
 - ensure that all players are valid
 - instantiate players with chips
 - create initial hand

## Big Components

### Poker Engine, all the poker stuff
validation, who won, split pots, etc

### Identity/Auth/Connections
Infra for setting up lobbies/tables

### Admin/UI/Player Integration
Making it so humans can play with/against bots
