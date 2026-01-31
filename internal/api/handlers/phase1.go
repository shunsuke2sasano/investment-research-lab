package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"investment_committee/internal/domain"
	"investment_committee/internal/phase1/fetcher"
)

func (s *Server) HandlePhase1Runs(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) == 0 {
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var in RunInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		cfg, err := json.Marshal(in.Config)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid config")
			return
		}
		id, err := s.store.CreateRun(r.Context(), in.Mode, cfg)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		if err := maybeFetchPhase1Docs(r.Context(), s, id, in.Config); err != nil {
			// fetch errors are non-fatal in v1
		}
		WriteJSON(w, http.StatusOK, map[string]string{"run_id": id})
		return
	}

	runID := rest[0]
	if len(rest) == 1 {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		run, err := s.store.GetRun(r.Context(), runID)
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, run)
		return
	}

	if len(rest) == 2 && rest[1] == "events" {
		switch r.Method {
		case http.MethodGet:
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			events, cursor, err := s.store.ListPhase1RunEventsByRunID(r.Context(), runID, limit, r.URL.Query().Get("cursor"))
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "list failed")
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{
				"items":       events,
				"next_cursor": cursor,
			})
			return
		case http.MethodPost:
			var in RunEventInput
			if err := DecodeJSON(r, &in); err != nil {
				WriteError(w, http.StatusBadRequest, "invalid json")
				return
			}
			if in.EventType == "" {
				WriteError(w, http.StatusBadRequest, "event_type required")
				return
			}
			if !domain.IsAllowedPhase1EventType(in.EventType) {
				WriteError(w, http.StatusBadRequest, "unknown event_type")
				return
			}
			seq, err := s.store.AppendEventToRun(r.Context(), runID, in)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "create failed")
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"run_id": runID, "seq": seq})
			return
		default:
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	if len(rest) == 2 && rest[1] == "anomaly-summary" {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		summary, err := s.store.GetAnomalySummaryByRun(r.Context(), runID)
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, summary)
		return
	}

	if len(rest) == 2 && rest[1] == "trigger-decision" {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		td, err := s.store.GetTriggerDecisionByRun(r.Context(), runID)
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, td)
		return
	}

	if len(rest) == 2 && rest[1] == "handoffs" {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h, err := s.store.ListHandoffsByRun(r.Context(), runID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "list failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": h})
		return
	}

	WriteError(w, http.StatusNotFound, "not found")
}

func maybeFetchPhase1Docs(ctx Context, s *Server, runID string, rawCfg map[string]any) error {
	cfg, err := fetcher.ParseConfig(rawCfg)
	if err != nil {
		return err
	}
	if len(cfg.Sources) == 0 {
		return nil
	}
	reg := fetcher.NewRegistry()
	reg.Register(fetcher.StubIRFetcher{})
	reg.Register(fetcher.StubSECFetcher{})
	reg.Register(fetcher.StubEDINETFetcher{})

	for _, src := range cfg.Sources {
		f, ok := reg.Get(src)
		if !ok {
			continue
		}
		docs, err := f.Fetch(ctx, cfg)
		if err != nil {
			continue
		}
		payload := map[string]any{
			"source":    src,
			"documents": docsToPayload(docs),
		}
		_, _ = s.store.AppendEventToRun(ctx, runID, RunEventInput{
			EventType: domain.Phase1EventDocFetched,
			Source:    domain.Phase1EventSourceOther,
			Payload:   payload,
		})
	}
	return nil
}

func docsToPayload(docs []fetcher.Document) []map[string]any {
	out := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		out = append(out, map[string]any{
			"doc_id":       d.DocID,
			"title":        d.Title,
			"url":          d.URL,
			"published_at": d.PublishedAt.UTC().Format(time.RFC3339),
			"ticker":       d.Ticker,
			"summary":      d.Summary,
		})
	}
	return out
}
