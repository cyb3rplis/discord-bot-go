#!/bin/bash

basedir=$(dirname "$0")/data/sounds
printf "walking directory %s\n\n" "$basedir"

for dir in ./data/sounds/*; do
  if [ -d "$dir" ]; then
    if [ ! -d "$dir" ]; then
      continue
    fi
    printf "walking directory %s\n" "$dir"
    for f in "$dir"/*.mp3.done; do
      if [ ! -f "$f" ]; then
        printf "skipping %s\n" "$dir"
        continue
      fi
      printf "renaming %s to %s\n" "$(basename "$f")" "$(basename "$f" .done)"
      mv "$f" "$dir/$(basename "$f" .done)"
    done
  fi
done