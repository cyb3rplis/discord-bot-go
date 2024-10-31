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
- [x] Play music from youtube
- [x] Text2Speech
- [x] Statistics (Sounds, User)
- [x] Most used sounds (Top 10, Top 20)
- [x] Add, Remove and List Favorites
- [ ] Playlists
- [ ] Volume Control (might only be possible when converting audio)
- [ ] Pin New Sounds as Message (with buttons?)
- [ ] Delete sounds (admins only)
- [ ] Move sounds (admins only, e.g. ".move fart2 custom")
- [ ] Add sounds via yt-dlp (admins only, ".save https:// soundName categoryName \<start\> \<end\>")
- [ ] Jail users for a specified amount of time (cant use sounds)
- [ ] Pull admin users from roles on server (currently hardcored)

## Executing the bot

Place all your sound files in DCA format in `./data/sounds`.
They should all be within subfolders, which act as categories for the sound bot.

```
$ ls -lR ./data/sounds
./data/sounds
total 4
drwxr-xr-x 2 user user 4096 Oct 30 18:56 test

./data/sounds/test:
total 4
-rw-r--r-- 1 user user 5 Oct 30 18:56 file.dca
```

IMPORTANT: the `./data` directory will also contain the `soundbot.db` database, so make sure to create a backup of the folder regularly.

Make sure to place your token in the `.env` file in the root directory of the project:

```
DISCORD_BOT_TOKEN=your_token_here
```

Run the following command to build the docker container (in the future hosted on github, modified `compose.yml` to pull up to date image):

```
$ docker build -t discord-bot-go .
```

Run the container:

```
$ docker compose up
```

Or directly send it to the background:

```
$ docker compose up -d
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
$ docker compose down
```

## Development

Run the container like so:

```
$ docker compose -f compose.dev.yml up --watch
```

This will rebuild the image once any changes are made to the backend.
