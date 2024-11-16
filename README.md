# Discord Bot

This is an implementation of a Discord bot using Go that I'm doing as a learning exercise and a hobby.

## Current capabilities

- Perform the handshake and connect to Discord Event Gateway
- Maintain connection sending heartbeats
  - For now it is unable to reconnect if something fails

## Requirements

- Go: `v1.23`
- Discord account

## Configuration and Running

1. First step is to create a bot account and get a token to be able to authenticate the bot. Follow the [discord.py tutorial](https://discordpy.readthedocs.io/en/stable/discord.html)

2. Setup the `.env` file by using `cp .env.example .env` and inserting your discord token

3. Then run `go mod tidy` and `go run .`
