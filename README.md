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
- [ ] Volume Control
- [ ] Pin New Sounds as Message (with buttons?)
- [ ] Delete sounds (admins only)
- [ ] Move sounds (admins only, e.g. ".move fart2 custom")
- [ ] Add sounds via yt-dlp (admins only, ".save https:// soundName categoryName <start> <end>")
- [ ] Jail users for a specified amount of time (cant use sounds)

# 1 Getting Started

## 1.1 Local Test environment

### 1.1.1 Config File and go

Copy `config.json` to `config.local.json` and make any modifications you wish not to commit.

Make sure your go paths are set correctly (arch example):

```
export GOROOT=/usr/lib/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin
export GO111MODULE=on
```

### 1.1.2 Dependencies

Below dependencies need to be met to make the bot work.

#### 1.1.2.1 sqlite3

```
sudo pacman -S sqlite3
```

#### 1.1.2.2 ffmpeg

Install ffmpeg for sound conversion

Example Command to convert mp3 to dca:

```
ffmpeg -i test.mp3 -filter:a "loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1" -f s16le -ar 48000 -ac 2 pipe:1 | dca > test.dca
```

#### 1.1.2.3 dca

Install dca for decoding audio files

```
go install github.com/bwmarrin/dca/cmd/dca@latest
```

Make sure that the `dca` tool is working on your OS, if not it needs manual compiling.

#### 1.1.2.4 piper

Install piper for text2speech

https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz

Make sure to place the above files in the `./tts` folder and have the voice model and config installed.

```
echo 'Deine Mutter ist so fett, sie piepst beim rückwärtsgehen' | ./piper \
  --model de_DE-thorsten-medium.onnx \
  --output_file welcome.wav
```

#### 1.1.2.5 yt-dlp

For youtube support, have yt-dlp. Download somewhere locally and specify in config.json

```
https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp
```

## 2. Build

```
./build.sh
```

## 3. Docker

Place all your sound files in DCA format in `incoming`. They should all be within subfolders, which act as categories for the sound bot.

```
$ ls -lR incoming
incoming:
total 4
drwxr-xr-x 2 user user 4096 Oct 30 18:56 test

incoming/test:
total 4
-rw-r--r-- 1 user user 5 Oct 30 18:56 file.dca
```

Run the following command to build the docker container:

```
docker build -t discord-bot-go -f Dockerfile.run .
```

Run the container:

```
docker run --rm discord-bot-go
```
