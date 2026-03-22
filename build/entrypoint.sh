#!/bin/sh
set -e

APP_MODE=${APP_MODE:-web}

case "$APP_MODE" in
    web)     exec /app/bin/web ;;
    *)       echo "Unknown APP_MODE: $APP_MODE"; exit 1 ;;
esac
