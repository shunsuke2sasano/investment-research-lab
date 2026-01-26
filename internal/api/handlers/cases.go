package handlers

import (
	"net/http"
	"strconv"
)

func (s *Server) HandleCases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var in CaseInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		id, err := s.store.CreateCase(r.Context(), in)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id})
	case http.MethodGet:
		q := r.URL.Query()
		var status *string
		if v := q.Get("status"); v != "" {
			status = &v
		}
		limit, _ := strconv.Atoi(q.Get("limit"))
		items, cursor, err := s.store.ListCases(r.Context(), CaseFilterInput{
			Status: status,
			Limit:  limit,
			Cursor: q.Get("cursor"),
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "list failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{
			"items":       items,
			"next_cursor": cursor,
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) HandleCase(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) == 1 && r.Method == http.MethodGet {
		detail, err := s.store.GetCaseDetail(r.Context(), rest[0])
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, detail)
		return
	}
	if len(rest) == 2 && rest[1] == "artifacts" {
		s.HandleArtifacts(w, r, rest[0])
		return
	}
	if len(rest) == 2 && rest[1] == "decisions" {
		s.HandleDecisions(w, r, rest[0])
		return
	}
	if len(rest) == 2 && rest[1] == "monitoring-plans" {
		s.HandleMonitoringPlans(w, r, rest[0])
		return
	}
	WriteError(w, http.StatusNotFound, "not found")
}
