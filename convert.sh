#!/bin/bash

basedir=$(dirname "$0")/sounds
printf "walking directory %s\n\n" "$basedir"
# convert all .mp3 files to .dca files
for f in ./sounds/*.mp3; do
  printf "converting %s to %s\n" "$(basename "$f")" "$(basename "$f" .mp3).dca"
  ffmpeg -i "$f" -hide_banner -loglevel error -filter:a "loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1" -f s16le -ar 48000 -ac 2 pipe:1 | dca > "$basedir/$(basename "$f" .mp3).dca"
done