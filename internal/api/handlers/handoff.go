package handlers

import (
	"net/http"
	"time"

	"investment_committee/internal/domain"
)

func (s *Server) HandleHandoffs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var in HandoffInput
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateHandoffPacket(in); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	events, _, err := s.store.ListPhase1RunEventsByRunID(r.Context(), in.RunID, 200, "")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list failed")
		return
	}
	if in.Packet == nil {
		in.Packet = map[string]any{}
	}
	phase1Packet := buildPhase1Packet(in.RunID, events)
	in.Packet["run_events"] = events
	in.Packet["phase1"] = phase1Packet
	in.Packet["version"] = 1
	in.Packet["phases"] = map[string]any{"phase1": phase1Packet}
	id, err := s.store.CreateHandoff(r.Context(), in)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "create failed")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id, "status": "created"})
}

func (s *Server) HandleHandoff(w http.ResponseWriter, r *http.Request, rest []string) {
	if len(rest) == 1 && r.Method == http.MethodGet {
		h, err := s.store.GetHandoff(r.Context(), rest[0])
		if err != nil {
			WriteError(w, http.StatusNotFound, "not found")
			return
		}
		WriteJSON(w, http.StatusOK, h)
		return
	}
	if len(rest) == 2 && rest[1] == "attach-case" && r.Method == http.MethodPost {
		var body struct {
			Case CaseInput `json:"case"`
		}
		if err := DecodeJSON(r, &body); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		id, err := s.store.AttachCaseToHandoff(r.Context(), rest[0], body.Case)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "attach failed")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"case_id": id})
		return
	}
	WriteError(w, http.StatusNotFound, "not found")
}

func validateHandoffPacket(in HandoffInput) error {
	if in.FromPhase != 1 {
		return errInvalid("from_phase must be 1")
	}
	// Phase3: IndustryReview, Phase5: Screening+Exploration
	if in.ToPhase != 3 && in.ToPhase != 5 {
		return errInvalid("to_phase must be 3 or 5")
	}
	if in.HandoffType != "light" && in.HandoffType != "heavy" {
		return errInvalid("handoff_type must be light or heavy")
	}
	if in.HandoffType == "light" && in.ToPhase != 5 {
		return errInvalid("light must go to phase 5")
	}
	if in.HandoffType == "heavy" && in.ToPhase != 3 {
		return errInvalid("heavy must go to phase 3")
	}

	packet := in.Packet
	required := []string{"handoff_type", "from_phase", "to_phase", "universe_item_ids", "event_ids", "trigger_decision_id", "created_at", "payload"}
	for _, k := range required {
		if _, ok := packet[k]; !ok {
			return errInvalid("packet missing " + k)
		}
	}
	if t, ok := packet["handoff_type"].(string); !ok || t != in.HandoffType {
		return errInvalid("packet.handoff_type mismatch")
	}
	if fp, ok := packet["from_phase"].(float64); !ok || int(fp) != in.FromPhase {
		return errInvalid("packet.from_phase mismatch")
	}
	if tp, ok := packet["to_phase"].(float64); !ok || int(tp) != in.ToPhase {
		return errInvalid("packet.to_phase mismatch")
	}
	if created, ok := packet["created_at"].(string); ok {
		if _, err := time.Parse(time.RFC3339, created); err != nil {
			return errInvalid("packet.created_at must be RFC3339")
		}
	} else {
		return errInvalid("packet.created_at must be string")
	}
	payload, ok := packet["payload"].(map[string]any)
	if !ok {
		return errInvalid("packet.payload must be object")
	}
	if vRaw, ok := packet["version"]; ok {
		if !isJSONNumber(vRaw) {
			return errInvalid("packet.version must be number")
		}
	}
	if phasesRaw, ok := packet["phases"]; ok {
		phases, ok := phasesRaw.(map[string]any)
		if !ok {
			return errInvalid("packet.phases must be object")
		}
		if phase1Raw, ok := phases["phase1"]; ok {
			if err := validatePhase1Packet(phase1Raw, in.RunID); err != nil {
				return err
			}
		}
	}
	if phase1Raw, ok := packet["phase1"]; ok {
		if err := validatePhase1Packet(phase1Raw, in.RunID); err != nil {
			return err
		}
	}
	if in.HandoffType == "light" {
		if _, ok := payload["summary_md"].(string); !ok {
			return errInvalid("light.payload.summary_md required")
		}
		if _, ok := payload["hypothesis_seeds"].([]any); !ok {
			return errInvalid("light.payload.hypothesis_seeds required")
		}
		if _, ok := payload["key_metrics"].(map[string]any); !ok {
			return errInvalid("light.payload.key_metrics required")
		}
	}
	if in.HandoffType == "heavy" {
		if _, ok := payload["summary_md"].(string); !ok {
			return errInvalid("heavy.payload.summary_md required")
		}
		if _, ok := payload["industry_scope"].(string); !ok {
			return errInvalid("heavy.payload.industry_scope required")
		}
		if _, ok := payload["value_pool_notes"].(string); !ok {
			return errInvalid("heavy.payload.value_pool_notes required")
		}
		if _, ok := payload["key_questions"].([]any); !ok {
			return errInvalid("heavy.payload.key_questions required")
		}
	}

	return nil
}

type errInvalid string

func (e errInvalid) Error() string { return string(e) }

func buildPhase1Packet(runID string, events []Phase1RunEvent) map[string]any {
	inputs := make([]domain.Phase1EventProjectionInput, 0, len(events))
	for _, e := range events {
		inputs = append(inputs, domain.Phase1EventProjectionInput{
			EventType: e.EventType,
			Source:    e.Source,
			Seq:       e.Seq,
		})
	}
	meta := domain.ProjectPhase1Events(inputs)
	return map[string]any{
		"run_id": runID,
		"events": events,
		"meta":   meta,
	}
}

func validatePhase1Packet(raw any, runID string) error {
	phase1, ok := raw.(map[string]any)
	if !ok {
		return errInvalid("packet.phase1 must be object")
	}
	if runIDRaw, ok := phase1["run_id"]; ok {
		runIDStr, ok := runIDRaw.(string)
		if !ok {
			return errInvalid("packet.phase1.run_id must be string")
		}
		if runIDStr != runID {
			return errInvalid("packet.phase1.run_id mismatch")
		}
	}
	if eventsRaw, ok := phase1["events"]; ok {
		if _, ok := eventsRaw.([]any); !ok {
			if _, ok := eventsRaw.([]Phase1RunEvent); !ok {
				return errInvalid("packet.phase1.events must be array")
			}
		}
	}
	return nil
}

func isJSONNumber(v any) bool {
	switch v.(type) {
	case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return true
	default:
		return false
	}
}
