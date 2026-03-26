# discord-bot-go

<p align="center">
  <img src="./assets/discord-bot-logo2.webp" alt="discord-bot-go" width="350"/>
</p>

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white" alt="Go"></a>
  <a href="https://github.com/bwmarrin/discordgo"><img src="https://img.shields.io/badge/DiscordGo-v0.28.1-5865F2?logo=discord&logoColor=white" alt="DiscordGo"></a>
  <a href="https://hub.docker.com/r/cyb3rplis/discord-bot-go"><img src="https://img.shields.io/badge/Docker-Hub-2496ED?logo=docker&logoColor=white" alt="Docker Hub"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-MIT-green" alt="MIT License"></a>
</p>

A feature-rich [Discord](https://discord.com/) soundboard bot written in [Go](https://golang.org/) using the [DiscordGo](https://github.com/bwmarrin/discordgo/) library. Play and manage audio directly in Discord voice channels from local files or YouTube via [yt-dlp](https://github.com/yt-dlp/yt-dlp). Fully interactive with button-based controls — no commands needed.

## Features

- **Soundboard** — Organize and play sounds from categorized local MP3 files
- **YouTube playback** — Stream audio from video platforms via yt-dlp
- **Button controls** — Interactive UI with no command-based interactions required
- **Categories** — Group sounds into categories for easy browsing
- **Favorites** — Add, remove, and list personal favorites
- **Statistics** — View sound and user stats, including a top 10 most-played leaderboard
- **Sound management** — Add, delete, and move sounds; pin new sounds as messages
- **Gulag** — Temporarily restrict a user from using sounds for a specified duration

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.23 |
| Discord library | [DiscordGo](https://github.com/bwmarrin/discordgo/) |
| Database | SQLite ([go-sqlite3](https://github.com/mattn/go-sqlite3)) |
| Scheduler | [gocron](https://github.com/go-co-op/gocron) |
| Audio encoding | [DCA](https://github.com/cyb3rplis/dca), FFmpeg |
| Media download | [yt-dlp](https://github.com/yt-dlp/yt-dlp) |
| Deployment | Docker, Docker Compose |
| CI/CD | GitHub Actions |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- A [Discord bot token](https://discord.com/developers/applications)

## Quick Start

### 1. Configure environment variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|----------|-------------|---------|
| `TOKEN` | Discord bot token | *(required)* |
| `APP_PATH` | Path to the data directory | `./` |
| `ADMIN_ROLE` | Role name with elevated bot permissions | *(required)* |
| `BOT_TIMEOUT` | Minutes of inactivity before leaving voice | `10` |
| `BOT_CHANNEL` | Channel where the bot pins messages | *(required)* |

### 2. Add sound files

Place MP3 files inside subdirectories of `${APP_PATH}/data/sounds/`. Each subdirectory acts as a **category**.

```
data/sounds/
├── memes/
│   ├── airhorn.mp3
│   └── bruh.mp3
└── music/
    └── intro.mp3
```

### 3. Run the bot

```bash
docker compose up -d
```

### 4. View logs

```bash
docker logs luren-bot
```

> **Note:** The `./data` directory also contains the `soundbot.db` SQLite database. Back up this directory regularly.

## Updating

Pull the latest image and restart:

```bash
docker compose pull
docker compose up -d
```

## Stopping

```bash
docker compose stop
```

## Restart Policy

The default `compose.yml` uses `restart: unless-stopped`, meaning the container will automatically restart on reboot unless manually stopped. See the [Docker restart policy docs](https://docs.docker.com/config/containers/start-containers-automatically/) for other options.

## Development

Build and run with live reload on code changes:

```bash
docker compose -f compose.dev.yml up -d --watch --build
```

The container rebuilds automatically when changes are detected in the backend.

## License

This project is licensed under the [MIT License](./LICENSE).
