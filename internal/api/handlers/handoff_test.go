package handlers

import "testing"

func TestValidateHandoffPacketLight(t *testing.T) {
	in := HandoffInput{
		RunID:       "run",
		HandoffType: "light",
		FromPhase:   1,
		ToPhase:     5,
		Packet: map[string]any{
			"handoff_type":        "light",
			"from_phase":          float64(1),
			"to_phase":            float64(5),
			"universe_item_ids":   []any{"u1"},
			"event_ids":           []any{"e1"},
			"trigger_decision_id": "t1",
			"created_at":          "2026-01-24T00:00:00Z",
			"payload": map[string]any{
				"summary_md":       "s",
				"hypothesis_seeds": []any{"h"},
				"key_metrics":      map[string]any{"k": "v"},
			},
		},
	}
	if err := validateHandoffPacket(in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateHandoffPacketHeavy(t *testing.T) {
	in := HandoffInput{
		RunID:       "run",
		HandoffType: "heavy",
		FromPhase:   1,
		ToPhase:     3,
		Packet: map[string]any{
			"handoff_type":        "heavy",
			"from_phase":          float64(1),
			"to_phase":            float64(3),
			"universe_item_ids":   []any{"u1"},
			"event_ids":           []any{"e1"},
			"trigger_decision_id": "t1",
			"created_at":          "2026-01-24T00:00:00Z",
			"payload": map[string]any{
				"summary_md":       "s",
				"industry_scope":   "scope",
				"value_pool_notes": "notes",
				"key_questions":    []any{"q"},
			},
		},
	}
	if err := validateHandoffPacket(in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
