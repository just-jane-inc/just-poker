# using the sdk

## getting started

Setup the local python development environment:

- create a virtual env with `python -m venv .venv`
- setup configurations `mv config/.env.example config/.env`
- activate the virtual environment
  - Linux: `source .venv/bin/activate`
  - Windows: `.venv\Scripts\Activate.ps1`
- setup the toml project environment `pip install -e .`

# helper scripts

the environment includes the following executable scripts to help interact with the game server.
these CLI tools all contain a --help option that prints information about how to use them.

## `poker-tui`

a textual user interface that can be used to view/play a game from the perspective of a player.
the tui can be used to setup, join, play, or just observe games.
some notes to keep in mind when using the TUI:

- users with the admin role (e.g. jill) are not allowed to be _players_ in a game
- users joining a game after it has started do so as observers
- admins observing a game see all cards for all players
- the tui will often crash on invalid inputs, you can just rejoin the game

```
// setup a game between two players using the default game configuration and users
// acquired from config/test_users.csv. insert your own players in this file in order
// for the setup to use them, ideally place them first/early in the csv. games can
// support up to 8 players
poker-tui --setup 2

// joins the game 6969 with the user rae, assumes rae exists in config/test_users.csv.
poker-tui --game-id 6969 --user rae
```

## `new-game`

creates a game on the server that can be joined by users in order to play the game.
in order to create a game you must provided a configuration file, an example file can be
found in config/new_game_config/0.json

```
// creates a game that can support up to 4 players, uses 100/50 blinds, 
// and auto starts hand
new-game --config config/new_game_config/0.json
```

## `join-game`

join a created game with provided user(s).

```
join-game --users jane, rae, red, wolf --id 42069
```

## `listen-game`

attach a listener to a running game that echos websocket messages to stdout and a provided file.
can be useful for debugging events or just capturing game states. not that attaching a listener with an admin (jill) token will allow the logs to contain all cards.
if the --file argument is not provided the logfile will be created in the current working directory with a name that includes a timestamp.

```
listen-game --id 42069 --user jill --file game_logs/some_log_file
```

## `start-game`

start a created game - locking the ability to join and beginning play

## `get-game`

get the state of a running game

## jamble-bot

an example bot implementation
