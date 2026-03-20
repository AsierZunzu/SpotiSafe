package spotify

import "encoding/json"

// flexString unmarshals a JSON string as a string, and any other JSON type
// (number, null, bool) as an empty string. This works around Spotify API
// inconsistencies where fields like "previous" are sometimes 0 instead of null.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexString(s)
	}
	// number, null, bool — treat as empty
	return nil
}

// UserProfile represents the authenticated user's Spotify profile.
type UserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Country     string `json:"country"`
	Product     string `json:"product"`
	Followers   struct {
		Total int `json:"total"`
	} `json:"followers"`
	Images []Image `json:"images"`
}

// Image is a Spotify image object.
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// Artist is a simplified Spotify artist object.
type Artist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Genres     []string `json:"genres"`
	Popularity int      `json:"popularity"`
	Followers  struct {
		Total int `json:"total"`
	} `json:"followers"`
	Images       []Image           `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
}

// Album is a simplified Spotify album object.
type Album struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	AlbumType    string            `json:"album_type"`
	ReleaseDate  string            `json:"release_date"`
	TotalTracks  int               `json:"total_tracks"`
	Artists      []Artist          `json:"artists"`
	Images       []Image           `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
}

// Track is a simplified Spotify track object.
type Track struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	DurationMs   int               `json:"duration_ms"`
	Explicit     bool              `json:"explicit"`
	Popularity   int               `json:"popularity"`
	TrackNumber  int               `json:"track_number"`
	Artists      []Artist          `json:"artists"`
	Album        Album             `json:"album"`
	ExternalURLs map[string]string `json:"external_urls"`
}

// SavedTrack wraps a track with the time it was added to the library.
type SavedTrack struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}

// SavedAlbum wraps an album with the time it was saved.
type SavedAlbum struct {
	AddedAt string `json:"added_at"`
	Album   Album  `json:"album"`
}

// Episode is a Spotify podcast episode object.
type Episode struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	DurationMs   int               `json:"duration_ms"`
	Explicit     bool              `json:"explicit"`
	ReleaseDate  string            `json:"release_date"`
	Show         Show              `json:"show"`
	ExternalURLs map[string]string `json:"external_urls"`
}

// SavedEpisode wraps an episode with the time it was saved.
type SavedEpisode struct {
	AddedAt string  `json:"added_at"`
	Episode Episode `json:"episode"`
}

// Show is a Spotify podcast show object.
type Show struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Publisher     string            `json:"publisher"`
	Languages     []string          `json:"languages"`
	TotalEpisodes int               `json:"total_episodes"`
	Images        []Image           `json:"images"`
	ExternalURLs  map[string]string `json:"external_urls"`
}

// SavedShow wraps a show with the time it was saved.
type SavedShow struct {
	AddedAt string `json:"added_at"`
	Show    Show   `json:"show"`
}

// PlaylistSummary is a simplified playlist object from the user's list.
type PlaylistSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Public        bool   `json:"public"`
	Collaborative bool   `json:"collaborative"`
	SnapshotID    string `json:"snapshot_id"`
	Tracks        struct {
		Total int `json:"total"`
	} `json:"tracks"`
	Owner struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Images       []Image           `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
}

// PlaylistTrack is a track within a playlist, with playlist-specific metadata.
type PlaylistTrack struct {
	AddedAt string `json:"added_at"`
	AddedBy struct {
		ID string `json:"id"`
	} `json:"added_by"`
	IsLocal bool  `json:"is_local"`
	Track   Track `json:"track"`
}

// TopArtist extends Artist with ranking context.
type TopArtist = Artist

// TopTrack extends Track with ranking context.
type TopTrack = Track

// RecentlyPlayedItem represents a recently played track with play context.
type RecentlyPlayedItem struct {
	PlayedAt string `json:"played_at"`
	Track    Track  `json:"track"`
	Context  *struct {
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"context"`
}

// Audiobook is a Spotify audiobook object.
type Audiobook struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Authors     []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Narrators []struct {
		Name string `json:"name"`
	} `json:"narrators"`
	Publisher     string            `json:"publisher"`
	Languages     []string          `json:"languages"`
	TotalChapters int               `json:"total_chapters"`
	Images        []Image           `json:"images"`
	ExternalURLs  map[string]string `json:"external_urls"`
}

// SavedAudiobook wraps an audiobook with the time it was saved.
type SavedAudiobook struct {
	// The API returns audiobooks directly (not wrapped like tracks/albums)
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Authors     []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Narrators []struct {
		Name string `json:"name"`
	} `json:"narrators"`
	Publisher     string            `json:"publisher"`
	Languages     []string          `json:"languages"`
	TotalChapters int               `json:"total_chapters"`
	Images        []Image           `json:"images"`
	ExternalURLs  map[string]string `json:"external_urls"`
}

// --- Paginated response envelopes ---

// OffsetPage is a generic Spotify offset-paginated response.
type OffsetPage[T any] struct {
	Items    []T        `json:"items"`
	Total    int        `json:"total"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	Next     flexString `json:"next"`
	Previous flexString `json:"previous"`
}

// CursorPage is a generic Spotify cursor-paginated response.
type CursorPage[T any] struct {
	Items   []T `json:"items"`
	Total   int `json:"total"`
	Limit   int `json:"limit"`
	Cursors struct {
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"cursors"`
	Next string `json:"next"`
}

// FollowedArtistsResponse is the envelope for the followed artists endpoint.
type FollowedArtistsResponse struct {
	Artists CursorPage[Artist] `json:"artists"`
}
