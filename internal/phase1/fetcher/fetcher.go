package fetcher

import (
	"context"
	"encoding/json"
	"time"
)

type Document struct {
	DocID       string
	Title       string
	URL         string
	PublishedAt time.Time
	Ticker      string
	Summary     string
}

type Phase1FetchConfig struct {
	Sources           []string `json:"sources"`
	MaxItemsPerSource int      `json:"max_items_per_source"`
}

type DocumentFetcher interface {
	Source() string
	Fetch(ctx context.Context, cfg Phase1FetchConfig) ([]Document, error)
}

type Registry struct {
	fetchers map[string]DocumentFetcher
}

func NewRegistry() *Registry {
	return &Registry{fetchers: map[string]DocumentFetcher{}}
}

func (r *Registry) Register(f DocumentFetcher) {
	if f == nil {
		return
	}
	r.fetchers[f.Source()] = f
}

func (r *Registry) Get(source string) (DocumentFetcher, bool) {
	f, ok := r.fetchers[source]
	return f, ok
}

func ParseConfig(raw map[string]any) (Phase1FetchConfig, error) {
	var cfg Phase1FetchConfig
	if raw == nil {
		return cfg, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
