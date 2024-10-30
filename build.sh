#!/bin/bash
set -eu

DOCKER_BUILDKIT=1 docker build --output dist/ -f Dockerfile.build .