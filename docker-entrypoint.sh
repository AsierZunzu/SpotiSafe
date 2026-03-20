#!/bin/sh
set -e

PUID=${PUID:-1000}
PGID=${PGID:-1000}

addgroup -g "$PGID" spotisafe 2>/dev/null || true
adduser -D -H -G spotisafe -u "$PUID" spotisafe 2>/dev/null || true

exec su-exec spotisafe "$@"
