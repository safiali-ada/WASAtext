#!/bin/bash
# Opens a temporary node container for development
docker run -it --rm \
    -v "$(pwd):/app" \
    -w /app/webui \
    -p 5173:5173 \
    -p 4173:4173 \
    node:20 \
    /bin/bash
