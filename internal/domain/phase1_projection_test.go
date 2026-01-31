package domain

import "testing"

func TestProjectPhase1Events(t *testing.T) {
	cases := []struct {
		name   string
		input  []Phase1EventProjectionInput
		checks func(t *testing.T, p Phase1Projection)
	}{
		{
			name:  "empty",
			input: nil,
			checks: func(t *testing.T, p Phase1Projection) {
				if p.TotalEvents != 0 {
					t.Fatalf("total_events=%d", p.TotalEvents)
				}
				if p.LastSeq != 0 {
					t.Fatalf("last_seq=%d", p.LastSeq)
				}
				if p.FinalizedPresent {
					t.Fatalf("finalized_present=true")
				}
			},
		},
		{
			name: "doc fetched + finalized",
			input: []Phase1EventProjectionInput{
				{EventType: Phase1EventDocFetched, Source: Phase1EventSourceManual, Seq: 1},
				{EventType: Phase1EventRunFinalized, Source: Phase1EventSourceSystem, Seq: 2},
			},
			checks: func(t *testing.T, p Phase1Projection) {
				if p.DocFetchedCount != 1 {
					t.Fatalf("doc_fetched_count=%d", p.DocFetchedCount)
				}
				if !p.FinalizedPresent {
					t.Fatalf("finalized_present=false")
				}
				if p.LastSeq != 2 {
					t.Fatalf("last_seq=%d", p.LastSeq)
				}
			},
		},
		{
			name: "source normalized",
			input: []Phase1EventProjectionInput{
				{EventType: Phase1EventNoteAdded, Source: "unknown", Seq: 3},
			},
			checks: func(t *testing.T, p Phase1Projection) {
				if p.CountsBySource[Phase1EventSourceOther] != 1 {
					t.Fatalf("counts_by_source[other]=%d", p.CountsBySource[Phase1EventSourceOther])
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ProjectPhase1Events(tc.input)
			tc.checks(t, p)
		})
	}
}
