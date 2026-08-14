- need to supply JUST_POKER_PEPPER in a .env file in this directory (note all other environment variables can also be configured here)
- docker compose up

## docker image login
- you will need to provide a PAT (classic) with read:package scope to the following command
[let me google that for you](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
docker login ghcr.io -u <githubusername>

## making users
you can use this command to make users - assumes the default port - this is an example to make jill as an admin user, users should use the "bot" type generally
curl -X POST http://localhost:7821/user -d '{"display_name": "jill", "twitch_id": "69420", "user_type": "admin"}'

## poker db
access to the local db is done on port 5432 (unless otherwise configured)

## getting pytest to pass:
- you will need to ensure that the config/test_users.csv has bot users for jane, rae, red, and wolf.
- putting users in this file is also required for the TUI

