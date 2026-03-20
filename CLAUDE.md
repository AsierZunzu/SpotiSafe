# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o spotisafe ./cmd/spotisafe

# Lint
golangci-lint run

# Run (after building and sourcing .env)
source .env && ./spotisafe

# Docker
docker compose up
```

There are no tests in this codebase.

## Architecture

SpotiSafe is a Go CLI that backs up a Spotify account to JSON files via the Spotify Web API.

**Flow:** `main.go` → load config → OAuth auth → fetch profile → create output dir → run backup jobs via orchestrator → exit with code (0=success, 1=fatal, 2=all jobs failed).

### Key packages

- **`internal/auth/`** — OAuth 2.0 PKCE flow. Tries loading a saved token first (with auto-refresh), falls back to browser-based authorization. Token persisted to `output/.spotify_token.json`. 8 OAuth scopes are requested.

- **`internal/spotify/`** — Spotify API client with retry logic (max 3 retries, exponential backoff + jitter, respects `Retry-After`). `paginator.go` handles offset pagination generically. `endpoints.go` wraps all API calls.

- **`internal/backup/`** — `Job` interface + 11 concrete jobs (tracks, albums, episodes, shows, artists, playlists, top artists/tracks, recently played, audiobooks). `orchestrator.go` runs all jobs sequentially with per-job soft-fail.

- **`internal/writer/`** — Creates timestamped output directories (`2006-01-02T15-04-05`). Wraps all output in a metadata envelope (`category`, `fetched_at`, `count`, `user_id`). Uses atomic temp-file-then-rename writes. Playlist track files go into a `playlist_tracks/` subdirectory.

- **`internal/config/`** — Loads env vars; only `SPOTIFY_CLIENT_ID` is required.

### Output structure

Each run produces `output/{timestamp}/` with one JSON file per data type, plus `playlist_tracks/{id}_{name}.json` for each playlist. The token is stored at `output/.spotify_token.json` across runs.

### Configuration

Copy `.env.example` to `.env` and set `SPOTIFY_CLIENT_ID`. The redirect URI defaults to `http://localhost:8888/callback`. First run opens a browser for OAuth; subsequent runs reuse the saved token.
