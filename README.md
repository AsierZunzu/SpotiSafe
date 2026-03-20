# SpotiSafe

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
