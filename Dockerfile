# Stage 1: Build the Go application
FROM golang:1.23 AS backend
RUN go install github.com/bwmarrin/dca/cmd/dca@latest
ENV GOFLAGS="-mod=vendor"
WORKDIR /app/backend
COPY backend/ ./
RUN GOOS=linux GOARCH=amd64 go build -mod vendor -o /dist/discord-bot-go main/main.go

# Stage 2: Set up the runtime environment
FROM debian:bookworm AS discord-bot-go
WORKDIR /app

# Install dependencies and download Go
RUN apt-get update && apt-get install -y \
    python3 \
    wget \
    xz-utils \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp
RUN wget https://github.com/rhasspy/piper/releases/latest/download/piper_linux_x86_64.tar.gz \
    && tar -xzvf piper_linux_x86_64.tar.gz -C /app \
    && rm piper_linux_x86_64.tar.gz \
    && wget https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/de/de_DE/thorsten/medium/de_DE-thorsten-medium.onnx?download=true -O /app/piper/de_DE-thorsten-medium.onnx \
    && wget https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/de/de_DE/thorsten/medium/de_DE-thorsten-medium.onnx.json?download=true.json -O /app/piper/de_DE-thorsten-medium.onnx.json
RUN wget https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz \
    && tar -xf ffmpeg-release-amd64-static.tar.xz \
    && ffmpegpath=$(ls | grep -e "static$") \
    && mv "$ffmpegpath"/ffmpeg /usr/bin/ffmpeg \
    && rm ffmpeg-release-amd64-static.tar.xz \
    && chmod a+rx /usr/bin/ffmpeg

COPY --from=backend /dist/discord-bot-go ./
COPY --from=backend /go/bin/dca /usr/local/bin/dca
CMD ["./discord-bot-go"]
