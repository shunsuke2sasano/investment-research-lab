package domain

import "strings"

const (
	Phase1EventRunFinalized       = "run.finalized"
	Phase1EventDocFetched         = "doc.fetched"
	Phase1EventNoteAdded          = "note.added"
	Phase1EventSignalDetected     = "signal.detected"
	Phase1EventUniverseMemberAdded = "universe.member_added"
)

var AllowedPhase1EventTypes = map[string]struct{}{
	Phase1EventRunFinalized:       {},
	Phase1EventDocFetched:         {},
	Phase1EventNoteAdded:          {},
	Phase1EventSignalDetected:     {},
	Phase1EventUniverseMemberAdded: {},
}

func IsAllowedPhase1EventType(t string) bool {
	t = strings.TrimSpace(t)
	if t == "" {
		return false
	}
	_, ok := AllowedPhase1EventTypes[t]
	return ok
}

const (
	Phase1EventSourceManual = "manual"
	Phase1EventSourceSystem = "system"
	Phase1EventSourceSEC    = "sec"
	Phase1EventSourceEDINET = "edinet"
	Phase1EventSourceOther  = "other"
)

var AllowedPhase1EventSources = map[string]struct{}{
	Phase1EventSourceManual: {},
	Phase1EventSourceSystem: {},
	Phase1EventSourceSEC:    {},
	Phase1EventSourceEDINET: {},
	Phase1EventSourceOther:  {},
}

func NormalizePhase1EventSource(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return Phase1EventSourceManual
	}
	if _, ok := AllowedPhase1EventSources[s]; ok {
		return s
	}
	return Phase1EventSourceOther
}
