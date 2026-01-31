package domain

type Phase1EventProjectionInput struct {
	EventType string
	Source    string
	Seq       int
}

type Phase1Projection struct {
	TotalEvents     int            `json:"total_events"`
	CountsByType    map[string]int `json:"counts_by_type"`
	CountsBySource  map[string]int `json:"counts_by_source"`
	DocFetchedCount int            `json:"doc_fetched_count"`
	FinalizedPresent bool          `json:"finalized_present"`
	LastSeq         int            `json:"last_seq"`
}

func ProjectPhase1Events(events []Phase1EventProjectionInput) Phase1Projection {
	out := Phase1Projection{
		TotalEvents:      len(events),
		CountsByType:     map[string]int{},
		CountsBySource:   map[string]int{},
		DocFetchedCount:  0,
		FinalizedPresent: false,
		LastSeq:          0,
	}
	for _, e := range events {
		out.CountsByType[e.EventType]++
		src := NormalizePhase1EventSource(e.Source)
		out.CountsBySource[src]++
		if e.EventType == Phase1EventDocFetched {
			out.DocFetchedCount++
		}
		if e.EventType == Phase1EventRunFinalized {
			out.FinalizedPresent = true
		}
		if e.Seq > out.LastSeq {
			out.LastSeq = e.Seq
		}
	}
	return out
}
