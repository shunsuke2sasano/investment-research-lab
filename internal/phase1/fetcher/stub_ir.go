package fetcher

import (
	"context"
	"fmt"
	"time"
)

type StubIRFetcher struct{}

func (f StubIRFetcher) Source() string { return "ir" }

func (f StubIRFetcher) Fetch(ctx context.Context, cfg Phase1FetchConfig) ([]Document, error) {
	max := cfg.MaxItemsPerSource
	if max <= 0 {
		max = 1
	}
	out := make([]Document, 0, max)
	for i := 1; i <= max; i++ {
		out = append(out, Document{
			DocID:       fmt.Sprintf("stub-%03d", i),
			Title:       "Stub IR Document",
			URL:         "https://example.com/ir/001",
			PublishedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Ticker:      "DUMMY",
			Summary:     "stub",
		})
	}
	return out, nil
}
