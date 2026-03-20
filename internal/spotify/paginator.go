package spotify

import (
	"context"
	"fmt"
)

// FetchAllOffset fetches all items from an offset-paginated Spotify endpoint.
// urlBase must already contain a "?" with any required query params (e.g. "?market=US").
// limit and offset are appended automatically.
func FetchAllOffset[T any](ctx context.Context, c *Client, urlBase string, limit int) ([]T, error) {
	var all []T
	offset := 0

	for {
		url := fmt.Sprintf("%s&limit=%d&offset=%d", urlBase, limit, offset)
		var page OffsetPage[T]
		if err := c.get(ctx, url, &page); err != nil {
			return all, err
		}

		all = append(all, page.Items...)

		if page.Next == "" || len(page.Items) == 0 {
			break
		}
		offset += len(page.Items)
	}

	return all, nil
}
