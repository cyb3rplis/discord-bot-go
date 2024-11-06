#!/bin/bash
set -eu
source dist/data/.env #load .env file for local development

echo "PATH from .env: $APP_PATH"

export TOKEN
export APP_PATH

echo "Start Bot local.."
go run main/main.go