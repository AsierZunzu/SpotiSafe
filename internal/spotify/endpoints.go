package spotify

import "fmt"

// GetProfile returns the authenticated user's Spotify profile.
func (c *Client) GetProfile() (*UserProfile, error) {
	var profile UserProfile
	if err := c.get(baseURL+"/me", &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetSavedTracks returns all tracks saved in the user's library.
func (c *Client) GetSavedTracks() ([]SavedTrack, error) {
	return FetchAllOffset[SavedTrack](c, baseURL+"/me/tracks?", 50)
}

// GetSavedAlbums returns all albums saved in the user's library.
func (c *Client) GetSavedAlbums() ([]SavedAlbum, error) {
	return FetchAllOffset[SavedAlbum](c, baseURL+"/me/albums?", 50)
}

// GetSavedEpisodes returns all podcast episodes saved in the user's library.
func (c *Client) GetSavedEpisodes() ([]SavedEpisode, error) {
	return FetchAllOffset[SavedEpisode](c, baseURL+"/me/episodes?", 50)
}

// GetSavedShows returns all podcast shows saved in the user's library.
func (c *Client) GetSavedShows() ([]SavedShow, error) {
	return FetchAllOffset[SavedShow](c, baseURL+"/me/shows?", 50)
}

// GetPlaylists returns all playlists in the user's library (own + followed).
func (c *Client) GetPlaylists() ([]PlaylistSummary, error) {
	return FetchAllOffset[PlaylistSummary](c, baseURL+"/me/playlists?", 50)
}

// GetPlaylistTracks returns all tracks in a given playlist.
func (c *Client) GetPlaylistTracks(playlistID string) ([]PlaylistTrack, error) {
	urlBase := fmt.Sprintf("%s/playlists/%s/tracks?", baseURL, playlistID)
	return FetchAllOffset[PlaylistTrack](c, urlBase, 50)
}

// GetTopArtists returns the user's top artists for the given time range.
// Valid ranges: "short_term", "medium_term", "long_term".
func (c *Client) GetTopArtists(timeRange string) ([]TopArtist, error) {
	urlBase := fmt.Sprintf("%s/me/top/artists?time_range=%s&", baseURL, timeRange)
	return FetchAllOffset[TopArtist](c, urlBase, 50)
}

// GetTopTracks returns the user's top tracks for the given time range.
// Valid ranges: "short_term", "medium_term", "long_term".
func (c *Client) GetTopTracks(timeRange string) ([]TopTrack, error) {
	urlBase := fmt.Sprintf("%s/me/top/tracks?time_range=%s&", baseURL, timeRange)
	return FetchAllOffset[TopTrack](c, urlBase, 50)
}

// GetFollowedArtists returns all artists the user follows using cursor pagination.
func (c *Client) GetFollowedArtists() ([]Artist, error) {
	var all []Artist
	after := ""

	for {
		url := fmt.Sprintf("%s/me/following?type=artist&limit=50", baseURL)
		if after != "" {
			url += "&after=" + after
		}

		var resp FollowedArtistsResponse
		if err := c.get(url, &resp); err != nil {
			return all, err
		}

		page := resp.Artists
		all = append(all, page.Items...)

		if page.Next == "" || len(page.Items) == 0 {
			break
		}
		after = page.Cursors.After
		if after == "" {
			break
		}
	}

	return all, nil
}

// GetRecentlyPlayed returns the user's recently played tracks using cursor pagination.
func (c *Client) GetRecentlyPlayed() ([]RecentlyPlayedItem, error) {
	var all []RecentlyPlayedItem
	after := ""

	for {
		url := fmt.Sprintf("%s/me/player/recently-played?limit=50", baseURL)
		if after != "" {
			url += "&after=" + after
		}

		var page CursorPage[RecentlyPlayedItem]
		if err := c.get(url, &page); err != nil {
			return all, err
		}

		all = append(all, page.Items...)

		if page.Next == "" || len(page.Items) == 0 {
			break
		}
		after = page.Cursors.After
		if after == "" {
			break
		}
	}

	return all, nil
}

// GetAudiobooks returns all audiobooks saved in the user's library.
func (c *Client) GetAudiobooks() ([]SavedAudiobook, error) {
	return FetchAllOffset[SavedAudiobook](c, baseURL+"/me/audiobooks?", 50)
}
