# discord-bot-go

<p align="center">
  <img src="./assets/discord-bot-logo2.webp" alt="discord-go-bot's image" width="350p"/>
</p>

discord-bot-go is a versatile [Discord](https://discord.com/) bot written in [Go](https://golang.org/) using the [DiscordGo](https://github.com/bwmarrin/discordgo/) library. \
It allows users to play and manage audio directly from Discord, supporting local files, YouTube, and text-to-speech functionalities. \
The bot also features interactive buttons for a seamless user experience, eliminating the need for command-based interactions.

# Features

- [x] Categories
- [x] Sync and Play music from local files
- [x] Interact with buttons to play music (no need to type commands)
- [x] Play music from video platform via yt-dlp
- [x] Statistics (Sounds, User)
- [x] Most used sounds (Top 10, Top 20)
- [x] Add, Remove and List Favorites
- [x] Pin New Sounds as Message
- [x] Delete sounds
- [x] Move sounds
- [x] Add sounds via yt-dlp
- [x] Jail users for a specified amount of time (cant use sounds)

## Executing the bot

INFO: Setting ${APP_PATH} is optional. If not set, it will default to the current directory.

Place all your sound files in MP3 format in `${APP_PATH}/data/sounds`.
They all have to be within subdirectories, which act as categories for the sound bot.

```
$ ls -lR ${APP_PATH}/data/sounds
${APP_PATH}/data/sounds
total 4
drwxr-xr-x 2 user user 4096 Oct 30 18:56 test

${APP_PATH}/data/sounds/test:
total 4
-rw-r--r-- 1 user user 5 Oct 30 18:56 file.mp3
```

IMPORTANT: the `./data` directory will also contain the `soundbot.db` database, so make sure to create a backup of the folder regularly.

Make sure to place your token in the `.env` file in the root directory of the project. See `example.env`.

Run the container:

```
$ docker compose up -d
```

Check logs with `docker logs`

```
$ docker logs luren-bot
```

When running the container this way, it will automatically start on a reboot (unless it was manually stopped before), as configured in `compose.yml`.

```
[...]
    restart: unless-stopped
```

Potential values can be set to the following:

```
no (default): Does not automatically restart the container.
always: Always restarts the container if it stops or the Docker daemon restarts.
unless-stopped: Restarts the container unless it was manually stopped.
on-failure: Only restarts the container if it exits with a non-zero exit code.
```

To stop the container:

```
$ docker compose stop
```

To fetch the newest version of the docker image:

```
$ docker compose pull
```

## Development

Run the container like so (omit --watch if you don't want to rebuild the running container on changes):

```
$ docker compose -f compose.dev.yml up -d --watch --build
```

This will rebuild the image once any changes are made to the backend.
