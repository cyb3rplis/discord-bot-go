# discord-bot-go

copy `config.json` to `config.local.json` and make any modifications you wish not to commit.

### dca
use dca in folder sounds to convert mp3 or wav
```
ffmpeg -i saul.mp3 -f s16le -ar 48000 -ac 2 pipe:1 | ./dca > sounds/songs/saul.dca
ffmpeg -i tts.wav -f s16le -ar 48000 -ac 2 pipe:1 | ./dca > sounds/tts/tts.dca
```

### using piper for tts
https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz

Make sure to place the above files in the ./tts folder and have the voice model and config installed.

```
echo 'Deine Mutter ist so fett, sie piepst beim rückwärtsgehen' | ./piper \
  --model de_DE-thorsten-medium.onnx \
  --output_file welcome.wav
```

### using yt-dlp for youtube
you need to have yt-dlp in your path.
The best way to do this is run the following before starting the bot:

```
$ python3 -m venv .venv
$ source .venv/bin/activate
$ pip install yt-dlp
$ cd main
$ go run main.go
```

Make sure that the `dca` tool is working on your OS, if not it needs manual compiling.

### To-Do

- control volume through config file values 0-100
