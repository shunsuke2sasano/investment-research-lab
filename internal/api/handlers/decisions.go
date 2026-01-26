package handlers

import "net/http"

func (s *Server) HandleDecisions(w http.ResponseWriter, r *http.Request, caseID string) {
	switch r.Method {
	case http.MethodPost:
		var in DecisionInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		in.CaseID = caseID
		id, err := s.store.CreateDecision(r.Context(), in)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id})
	case http.MethodGet:
		items, err := s.store.ListDecisionsByCase(r.Context(), caseID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "list failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) HandleDecision(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) == 1 && r.Method == http.MethodGet {
		d, err := s.store.GetDecision(r.Context(), rest[0])
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, d)
		return
	}
	WriteError(w, http.StatusNotFound, "not found")
}
