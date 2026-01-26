package handlers

import (
	"net/http"
	"strconv"
)

func (s *Server) HandleArtifacts(w http.ResponseWriter, r *http.Request, caseID string) {
	switch r.Method {
	case http.MethodPost:
		var in ArtifactInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		in.CaseID = caseID
		id, err := s.store.CreateArtifact(r.Context(), in)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id})
	case http.MethodGet:
		q := r.URL.Query()
		var phase *int
		if v := q.Get("phase"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "invalid phase")
				return
			}
			phase = &i
		}
		latest := q.Get("latest") == "true"
		items, err := s.store.ListArtifacts(r.Context(), caseID, ArtifactFilterInput{
			Phase:  phase,
			Latest: latest,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "list failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
