# SpotiSafe

> **Personal project** — built for self-use and as a testbed for exploring different AI-assisted programming techniques and tools. Shared as-is with no guarantees of stability, support, or general-purpose usability.

A Go CLI tool that backs up all your Spotify account data to local JSON files. Run it periodically to keep a snapshot of your library, playlists, listening history, and profile — for recovery or migration.

## What gets backed up

| File | Contents |
|---|---|
| `profile.json` | User profile |
| `saved_tracks.json` | Liked songs |
| `saved_albums.json` | Saved albums |
| `saved_episodes.json` | Saved podcast episodes |
| `saved_shows.json` | Followed podcasts |
| `followed_artists.json` | Followed artists |
| `playlists.json` | Playlist index |
| `playlist_tracks/{id}_{name}.json` | Full track list per playlist |
| `top_artists_{short,medium,long}_term.json` | Top artists by time range |
| `top_tracks_{short,medium,long}_term.json` | Top tracks by time range |
| `recently_played.json` | Recently played tracks |
| `audiobooks.json` | Saved audiobooks |

Each file is wrapped in a metadata envelope:
```json
{
  "metadata": { "category": "...", "fetched_at": "...", "count": 842, "user_id": "..." },
  "data": [...]
}
```

## Quick start

### 1. Create a Spotify app

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new app
3. Add `http://localhost:8888/callback` as a Redirect URI
4. Copy your **Client ID**

### 2. Configure

```bash
cp .env.example .env
# Edit .env and set SPOTIFY_CLIENT_ID=your_client_id
```

### 3. Run with Docker Compose

```bash
docker compose up
```

SpotiSafe will print an authorization URL. Open it in your browser, authorize the app, and the backup will start automatically.

Output is written to `./output/<timestamp>/`.

### 4. Subsequent runs

The OAuth token is saved to `./output/.spotify_token.json`. On re-runs, SpotiSafe uses it automatically (with auto-refresh) — no browser required.

```bash
docker compose up
```

## Building

### Docker image

```bash
docker build -t spotisafe:latest .
```

### Local binary

Requires Go 1.25+.

```bash
go build -o spotisafe ./cmd/spotisafe
```

To produce a smaller, fully static binary:

```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o spotisafe ./cmd/spotisafe
```

## Testing

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

Run a specific package:

```bash
go test ./internal/spotify/...
```

### Coverage

Print per-function coverage summary:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Open an interactive HTML report in the browser:

```bash
go tool cover -html=coverage.out
```

## Linting

Requires [golangci-lint](https://golangci-lint.run/welcome/install/).

```bash
golangci-lint run
```

## Run locally (without Docker)

```bash
go build -o spotisafe ./cmd/spotisafe
source .env
./spotisafe
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `SPOTIFY_CLIENT_ID` | *(required)* | Spotify application client ID |
| `SPOTIFY_CLIENT_SECRET` | *(empty)* | Not required for PKCE flow |
| `SPOTIFY_REDIRECT_URI` | `http://localhost:8888/callback` | OAuth redirect URI |
| `SPOTIFY_CALLBACK_PORT` | `8888` | Local callback server port |
| `SPOTIFY_OUTPUT_DIR` | `./output` | Where to write backup files |
| `SPOTIFY_TOKEN_FILE` | `./output/.spotify_token.json` | Token persistence path |
| `SPOTIFY_SCHEDULE` | *(unset)* | Cron expression for recurring backups (e.g. `0 2 * * *`). Unset = run once and exit |
| `SPOTIFY_SKIP_INITIAL_RUN` | `false` | Set to `true` to skip the immediate run at startup when a schedule is configured |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |

## Output structure

```
output/
└── 2026-03-20T14-05-00/
    ├── profile.json
    ├── saved_tracks.json
    ├── saved_albums.json
    ├── saved_episodes.json
    ├── saved_shows.json
    ├── followed_artists.json
    ├── playlists.json
    ├── playlist_tracks/
    │   └── {playlist_id}_{name}.json
    ├── top_artists_short_term.json
    ├── top_artists_medium_term.json
    ├── top_artists_long_term.json
    ├── top_tracks_short_term.json
    ├── top_tracks_medium_term.json
    ├── top_tracks_long_term.json
    ├── recently_played.json
    └── audiobooks.json
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | All jobs succeeded |
| `1` | Fatal error (config or auth) |
| `2` | All jobs failed |
