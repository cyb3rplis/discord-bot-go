# discord-bot-go

use dca in folder sounds to convert mp3 or wav
ffmpeg -i saul.mp3 -f s16le -ar 48000 -ac 2 pipe:1 | ./dca > saul.dca

# using piper for tts
https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz

```
echo 'Deine Mutter ist so fett, sie piepst beim rückwärtsgehen' | ./piper \
  --model de_DE-thorsten-medium.onnx \
  --output_file welcome.wav
```

# To-Do

- control volume through config file values 0-100
