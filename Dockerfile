# Stage 1: Build the Go application
FROM golang:1.23 AS backend
RUN go install github.com/bwmarrin/dca/cmd/dca@latest
ENV GOFLAGS="-mod=vendor"
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /app/backend
COPY backend/ ./
RUN go vet main/main.go
RUN go build -o /dist/discord-bot-go main/main.go

# Stage 2: Set up the runtime environment
FROM debian:bookworm-slim AS discord-bot-go
WORKDIR /app

# Install dependencies and download Go
RUN apt-get update && apt-get install -y \
    python3 \
    wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

COPY --from=mwader/static-ffmpeg:latest /ffmpeg /usr/local/bin/
COPY --from=backend /dist/discord-bot-go ./
COPY --from=backend /go/bin/dca /usr/local/bin/
CMD ["./discord-bot-go"]
