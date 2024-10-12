# discord-bot-go

use dca in folder sounds to convert mp3
ffmpeg -i saul.mp3 -f s16le -ar 48000 -ac 2 pipe:1 | ./dca > saul.dca

# To-Do

- control volume through config file values 0-100
