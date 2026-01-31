package fetcher

import (
	"context"
	"fmt"
	"time"
)

type StubSECFetcher struct{}

func (f StubSECFetcher) Source() string { return "sec" }

func (f StubSECFetcher) Fetch(ctx context.Context, cfg Phase1FetchConfig) ([]Document, error) {
	max := cfg.MaxItemsPerSource
	if max <= 0 {
		max = 1
	}
	out := make([]Document, 0, max)
	for i := 1; i <= max; i++ {
		out = append(out, Document{
			DocID:       fmt.Sprintf("sec-stub-%03d", i),
			Title:       "Stub SEC Document",
			URL:         "https://example.com/sec/001",
			PublishedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			Ticker:      "DUMMY",
			Summary:     "stub",
		})
	}
	return out, nil
}
