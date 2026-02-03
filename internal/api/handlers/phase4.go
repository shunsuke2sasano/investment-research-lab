package handlers

import "net/http"

func (s *Server) HandlePhase4Runs(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) != 0 {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var in Phase4RunInput
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

	phase3RunID := ""
	phase3PositioningPresent := false
	if raw, ok := phases["phase3"]; ok {
		if m, ok := raw.(map[string]any); ok {
			if v, ok := m["run_id"].(string); ok {
				phase3RunID = v
			}
			if p, ok := m["positioning"].(map[string]any); ok && p != nil {
				phase3PositioningPresent = true
			}
		}
	}

	phase2RunID := ""
	phase2CandidatesCount := 0
	phase2TemplatePresent := false

	if raw, ok := phases["phase2"]; ok {
		if m, ok := raw.(map[string]any); ok {
			if v, ok := m["run_id"].(string); ok {
				phase2RunID = v
			}
			if rawCandidates, ok := m["industry_candidates"]; ok {
				if arr, ok := rawCandidates.([]any); ok {
					phase2CandidatesCount = len(arr)
					for _, c := range arr {
						if m, ok := c.(map[string]any); ok {
							if v, ok := m["industry_id"].(string); ok && v == "__unset__" {
								phase2TemplatePresent = true
								break
							}
						}
					}
				}
			}
		}
	}

	runID, err := s.store.CreatePhase4Run(r.Context(), packet)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "create failed")
		return
	}

	meta := map[string]any{
		"source_phase3_run_id":            phase3RunID,
		"phase3_positioning_present":      phase3PositioningPresent,
		"source_phase2_run_id":            phase2RunID,
		"phase2_industry_candidates_count": phase2CandidatesCount,
		"phase2_template_present":         phase2TemplatePresent,
	}

	researchPlan := map[string]any{
		"key_questions": []any{},
		"hypotheses":    []any{},
		"info_needs":    []any{},
		"sources": map[string]any{
			"primary":   []any{},
			"secondary": []any{},
			"data":      []any{},
		},
		"artifacts": map[string]any{
			"notes_template_md": "",
			"checklist":         []any{},
		},
	}

	phase4 := map[string]any{
		"run_id":         runID,
		"research_plan":  researchPlan,
		"notes":          []any{},
		"meta":           meta,
	}
	phases["phase4"] = phase4
	packet["phases"] = phases
	if _, ok := packet["version"]; !ok {
		packet["version"] = float64(1)
	}

	if err := s.store.UpdatePhase4RunPacket(r.Context(), runID, packet); err != nil {
		WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}

	WriteJSON(w, http.StatusCreated, Phase4RunOutput{
		RunID:  runID,
		Packet: packet,
	})
}
