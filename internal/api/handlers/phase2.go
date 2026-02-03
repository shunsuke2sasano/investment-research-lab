package handlers

import "net/http"

func (s *Server) HandlePhase2Runs(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) != 0 {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var in Phase2RunInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if in.Packet == nil {
		WriteError(w, http.StatusBadRequest, "packet required")
		return
	}

	packet := in.Packet
	if v, ok := packet["version"]; ok {
		if !isVersionOne(v) {
			WriteError(w, http.StatusBadRequest, "packet.version must be number 1")
			return
		}
	}

	phases := map[string]any{}
	if raw, ok := packet["phases"]; ok {
		m, ok := raw.(map[string]any)
		if !ok {
			WriteError(w, http.StatusBadRequest, "packet.phases must be object")
			return
		}
		phases = m
	}

	phase1RunID := ""
	phase1TotalEvents := 0
	phase1LastSeq := 0
	phase1FinalizedPresent := false

	if raw, ok := phases["phase1"]; ok {
		if m, ok := raw.(map[string]any); ok {
			if v, ok := m["run_id"].(string); ok {
				phase1RunID = v
			}
			if metaRaw, ok := m["meta"]; ok {
				if meta, ok := metaRaw.(map[string]any); ok {
					if v, ok := toInt(meta["total_events"]); ok {
						phase1TotalEvents = v
					}
					if v, ok := toInt(meta["last_seq"]); ok {
						phase1LastSeq = v
					}
					if v, ok := meta["finalized_present"].(bool); ok {
						phase1FinalizedPresent = v
					}
				}
			}
		}
	}

	runID, err := s.store.CreatePhase2Run(r.Context(), packet)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "create failed")
		return
	}

	meta := map[string]any{
		"source_phase1_run_id":     phase1RunID,
		"phase1_total_events":      phase1TotalEvents,
		"phase1_last_seq":          phase1LastSeq,
		"phase1_finalized_present": phase1FinalizedPresent,
	}

	industryCandidates := []any{
		map[string]any{
			"industry_id": "__unset__",
			"source":      "system",
			"derived_from": map[string]any{
				"phase1_run_id": phase1RunID,
				"event_refs":    []any{},
			},
			"notes":      []any{},
			"confidence": nil,
		},
	}

	phase2 := map[string]any{
		"run_id":              runID,
		"industry_candidates": industryCandidates,
		"notes":               []any{},
		"meta":                meta,
	}
	phases["phase2"] = phase2
	packet["phases"] = phases
	if _, ok := packet["version"]; !ok {
		packet["version"] = 1
	}

	if err := s.store.UpdatePhase2RunPacket(r.Context(), runID, packet); err != nil {
		// non-fatal: return created run with packet
	}

	WriteJSON(w, http.StatusCreated, Phase2RunOutput{
		RunID:  runID,
		Packet: packet,
	})
}

func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case float32:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	case int32:
		return int(t), true
	case int16:
		return int(t), true
	case int8:
		return int(t), true
	case uint:
		return int(t), true
	case uint64:
		return int(t), true
	case uint32:
		return int(t), true
	case uint16:
		return int(t), true
	case uint8:
		return int(t), true
	default:
		return 0, false
	}
}
