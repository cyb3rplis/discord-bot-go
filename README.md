# discord-bot-go

<p align="center">
  <img src="./assets/discord-bot-logo2.webp" alt="discord-go-bot's image" width="350p"/>
</p>

discord-bot-go is a versatile [Discord](https://discord.com/) bot written in [Go](https://golang.org/) using the [DiscordGo](https://github.com/bwmarrin/discordgo/) library. \
It allows users to play and manage audio directly from Discord, supporting local files, YouTube, and text-to-speech functionalities. \
The bot also features interactive buttons for a seamless user experience, eliminating the need for command-based interactions.

## Features

- [x] Categories
- [x] Sync and Play music from local files
- [x] Interact with buttons to play music (no need to type commands)
- [x] Play music from youtube
- [x] Text2Speech
- [x] Statistics (Sounds, User)
- [x] Most used sounds (Top 10, Top 20)
- [x] Add, Remove and List Favorites
- [ ] Playlists
- [ ] Volume Control

## Getting Started

### Install

Copy `config.json` to `config.local.json` and make any modifications you wish not to commit.

Make sure your go paths are set correctly (arch example):

```
export GOROOT=/usr/lib/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin
export GO111MODULE=on
```

## Dependencies

Below dependencies need to be met to make the bot work.

### sqlite3

```
sudo pacman -S sqlite3
```

### ffmpeg

Install ffmpeg for sound conversion

Example Command to convert mp3 to dca:

```
ffmpeg -i test.mp3 -filter:a "loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1" -f s16le -ar 48000 -ac 2 pipe:1 | dca > test.dca
```

### dca

Install dca for decoding audio files

```
go install github.com/bwmarrin/dca/cmd/dca@latest
```

Make sure that the `dca` tool is working on your OS, if not it needs manual compiling.

### piper

Install piper for text2speech

https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz

Make sure to place the above files in the `./tts` folder and have the voice model and config installed.

```
echo 'Deine Mutter ist so fett, sie piepst beim rückwärtsgehen' | ./piper \
  --model de_DE-thorsten-medium.onnx \
  --output_file welcome.wav
```

### yt-dlp

For youtube support

you need to have yt-dlp in your path.
The best way to do this is run the following before starting the bot:

```
$ python3 -m venv .venv
$ source .venv/bin/activate
$ pip install yt-dlp
$ cd main
$ go run main.go
```

### Build

```
./build.sh
```

## Usage

Example usage of the bot here (screenshots, gifs, etc.):
