package spotify

import (
	"context"
	"fmt"
)

// GetProfile returns the authenticated user's Spotify profile.
func (c *Client) GetProfile(ctx context.Context) (*UserProfile, error) {
	var profile UserProfile
	if err := c.get(ctx, baseURL+"/me", &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetSavedTracks returns all tracks saved in the user's library.
func (c *Client) GetSavedTracks(ctx context.Context) ([]SavedTrack, error) {
	return FetchAllOffset[SavedTrack](ctx, c, baseURL+"/me/tracks?", 50)
}

// GetSavedAlbums returns all albums saved in the user's library.
func (c *Client) GetSavedAlbums(ctx context.Context) ([]SavedAlbum, error) {
	return FetchAllOffset[SavedAlbum](ctx, c, baseURL+"/me/albums?", 50)
}

// GetSavedEpisodes returns all podcast episodes saved in the user's library.
func (c *Client) GetSavedEpisodes(ctx context.Context) ([]SavedEpisode, error) {
	return FetchAllOffset[SavedEpisode](ctx, c, baseURL+"/me/episodes?", 50)
}

// GetSavedShows returns all podcast shows saved in the user's library.
func (c *Client) GetSavedShows(ctx context.Context) ([]SavedShow, error) {
	return FetchAllOffset[SavedShow](ctx, c, baseURL+"/me/shows?", 50)
}

// GetPlaylists returns all playlists in the user's library (own + followed).
func (c *Client) GetPlaylists(ctx context.Context) ([]PlaylistSummary, error) {
	return FetchAllOffset[PlaylistSummary](ctx, c, baseURL+"/me/playlists?", 50)
}

// GetPlaylistTracks returns all tracks in a given playlist.
func (c *Client) GetPlaylistTracks(ctx context.Context, playlistID string) ([]PlaylistTrack, error) {
	urlBase := fmt.Sprintf("%s/playlists/%s/tracks?", baseURL, playlistID)
	return FetchAllOffset[PlaylistTrack](ctx, c, urlBase, 50)
}

// GetTopArtists returns the user's top artists for the given time range.
// Valid ranges: "short_term", "medium_term", "long_term".
func (c *Client) GetTopArtists(ctx context.Context, timeRange string) ([]TopArtist, error) {
	urlBase := fmt.Sprintf("%s/me/top/artists?time_range=%s&", baseURL, timeRange)
	return FetchAllOffset[TopArtist](ctx, c, urlBase, 50)
}

// GetTopTracks returns the user's top tracks for the given time range.
// Valid ranges: "short_term", "medium_term", "long_term".
func (c *Client) GetTopTracks(ctx context.Context, timeRange string) ([]TopTrack, error) {
	urlBase := fmt.Sprintf("%s/me/top/tracks?time_range=%s&", baseURL, timeRange)
	return FetchAllOffset[TopTrack](ctx, c, urlBase, 50)
}

// GetFollowedArtists returns all artists the user follows using cursor pagination.
func (c *Client) GetFollowedArtists(ctx context.Context) ([]Artist, error) {
	var all []Artist
	after := ""

	for {
		url := fmt.Sprintf("%s/me/following?type=artist&limit=50", baseURL)
		if after != "" {
			url += "&after=" + after
		}

		var resp FollowedArtistsResponse
		if err := c.get(ctx, url, &resp); err != nil {
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
func (c *Client) GetRecentlyPlayed(ctx context.Context) ([]RecentlyPlayedItem, error) {
	var all []RecentlyPlayedItem
	after := ""

	for {
		url := fmt.Sprintf("%s/me/player/recently-played?limit=50", baseURL)
		if after != "" {
			url += "&after=" + after
		}

		var page CursorPage[RecentlyPlayedItem]
		if err := c.get(ctx, url, &page); err != nil {
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
func (c *Client) GetAudiobooks(ctx context.Context) ([]SavedAudiobook, error) {
	return FetchAllOffset[SavedAudiobook](ctx, c, baseURL+"/me/audiobooks?", 50)
}
