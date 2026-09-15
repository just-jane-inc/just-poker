# using the sdk

## getting started

Setup the local python development environment:

- create a virtual env with `python -m venv .venv`
- activate the virtual environment `source .venv/bin/activate`
- setup the toml project environment `pip install -e .`

# helper scripts

the environment includes the following executable scripts to help interact with the game server.
these CLI tools all contain a --help option that prints information about how to use them.

## `poker-tui`

a textual user interface that can be used to view/play a game from the perspective of a player

## `new-game`

creates a game on the server

## `join-game`

join a created game with a set of users

## `listen-game`

attach a listener to a running game that echos websocket messages to stdout

## `start-game`

start a created game - locking the ability to join and beginning play

## `get-game`

get the state of a running game

## jamble-bot

an example bot implementation
