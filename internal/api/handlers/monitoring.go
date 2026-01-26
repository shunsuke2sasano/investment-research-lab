package handlers

import (
	"net/http"
	"strconv"
)

func (s *Server) HandleMonitoringPlans(w http.ResponseWriter, r *http.Request, caseID string) {
	switch r.Method {
	case http.MethodPost:
		var in MonitoringPlanInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		in.CaseID = caseID
		id, err := s.store.CreateMonitoringPlan(r.Context(), in)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id, "status": "active"})
	case http.MethodGet:
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, cursor, err := s.store.ListMonitoringPlansByCase(r.Context(), caseID, limit, r.URL.Query().Get("cursor"))
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

func (s *Server) HandleMonitoringPlan(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) == 1 && r.Method == http.MethodGet {
		mp, err := s.store.GetMonitoringPlan(r.Context(), rest[0])
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, mp)
		return
	}
	if len(rest) == 2 && rest[1] == "alerts" {
		s.HandleAlerts(w, r, rest[0])
		return
	}
	WriteError(w, http.StatusNotFound, "not found")
}

func (s *Server) HandleAlerts(w http.ResponseWriter, r *http.Request, planID string) {
	switch r.Method {
	case http.MethodPost:
		var in AlertInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		in.MonitoringPlanID = planID
		id, err := s.store.CreateAlert(r.Context(), in)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id})
	case http.MethodGet:
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, cursor, err := s.store.ListAlertsByPlan(r.Context(), planID, limit, r.URL.Query().Get("cursor"))
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

func (s *Server) HandleAlert(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) == 2 && rest[1] == "ack" && r.Method == http.MethodPost {
		if err := s.store.AckAlert(r.Context(), rest[0]); err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	WriteError(w, http.StatusNotFound, "not found")
}
