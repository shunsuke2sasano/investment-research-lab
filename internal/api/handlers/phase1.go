package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
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
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		events, cursor, err := s.store.ListEventsByRun(r.Context(), runID, limit, r.URL.Query().Get("cursor"))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "list failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{
			"items":       events,
			"next_cursor": cursor,
		})
		return
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
