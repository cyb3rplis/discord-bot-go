# Stage 1: Build Backend
FROM golang:1.23 as backend
ENV GOFLAGS -mod=vendor
WORKDIR /build/backend
COPY go.mod go.sum main/main.go ./
COPY vendor vendor
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod vendor -o bin/discord-bot-go main/main.go

# Stage 2: Export content
FROM scratch AS scratch
COPY --from=backend /build/backend/bin/discord-bot-go ./
