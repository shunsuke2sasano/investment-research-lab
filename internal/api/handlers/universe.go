package handlers

import (
	"net/http"
	"strconv"
)

func (s *Server) HandleUniverseItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var in UniverseItemInput
		if err := DecodeJSON(r, &in); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		id, err := s.store.CreateUniverseItem(r.Context(), in)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "create failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id})
	case http.MethodGet:
		q := r.URL.Query()
		var active *bool
		if v := q.Get("active"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "invalid active")
				return
			}
			active = &b
		}
		var entityType *string
		if v := q.Get("entity_type"); v != "" {
			entityType = &v
		}
		limit, _ := strconv.Atoi(q.Get("limit"))

		items, cursor, err := s.store.ListUniverseItems(r.Context(), UniverseFilterInput{
			Active:     active,
			EntityType: entityType,
			Limit:      limit,
			Cursor:     q.Get("cursor"),
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

func (s *Server) HandleUniverseItem(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPatch {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var in UniverseUpdateInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.store.UpdateUniverseItem(r.Context(), id, in); err != nil {
		WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
