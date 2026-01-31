package fetcher

import (
	"context"
	"fmt"
	"time"
)

type StubEDINETFetcher struct{}

func (f StubEDINETFetcher) Source() string { return "edinet" }

func (f StubEDINETFetcher) Fetch(ctx context.Context, cfg Phase1FetchConfig) ([]Document, error) {
	max := cfg.MaxItemsPerSource
	if max <= 0 {
		max = 1
	}
	out := make([]Document, 0, max)
	for i := 1; i <= max; i++ {
		out = append(out, Document{
			DocID:       fmt.Sprintf("edinet-stub-%03d", i),
			Title:       "Stub EDINET Document",
			URL:         "https://example.com/edinet/001",
			PublishedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
			Ticker:      "DUMMY",
			Summary:     "stub",
		})
	}
	return out, nil
}
