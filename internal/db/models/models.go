package models

import (
	"encoding/json"
	"time"
)

type UniverseItem struct {
	ID         string          `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Name       string          `json:"name"`
	Keywords   json.RawMessage `json:"keywords"`
	Priority   int             `json:"priority"`
	IsActive   bool            `json:"is_active"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Run struct {
	ID         string          `json:"id"`
	Phase      int             `json:"phase"`
	Mode       string          `json:"mode"`
	Status     string          `json:"status"`
	ConfigJSON json.RawMessage `json:"config"`
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at,omitempty"`
	Error      *string         `json:"error,omitempty"`
}

type RawItem struct {
	ID         string     `json:"id"`
	RunID      string     `json:"run_id"`
	SourceType string     `json:"source_type"`
	SourceName *string    `json:"source_name,omitempty"`
	URL        string     `json:"url"`
	Title      string     `json:"title"`
	Published  *time.Time `json:"published_at,omitempty"`
	RawText    string     `json:"raw_text"`
	Hash       string     `json:"hash"`
	FetchedAt  time.Time  `json:"fetched_at"`
}

type Event struct {
	EventID    string          `json:"event_id"`
	RunID      string          `json:"run_id"`
	ObservedAt time.Time       `json:"observed_at"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Category   string          `json:"category"`
	Title      string          `json:"title"`
	FactsJSON  json.RawMessage `json:"facts_json"`
	ImpactJSON json.RawMessage `json:"impact_json,omitempty"`
	Sources    json.RawMessage `json:"sources_json"`
	Confidence float64         `json:"confidence"`
	DedupeKey  string          `json:"dedupe_key"`
	TagsJSON   json.RawMessage `json:"tags_json,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AnomalySummary struct {
	RunID     string          `json:"run_id"`
	Summary   json.RawMessage `json:"summary_json"`
	CreatedAt time.Time       `json:"created_at"`
}

type TriggerDecision struct {
	RunID     string          `json:"run_id"`
	Decision  json.RawMessage `json:"decision_json"`
	CreatedAt time.Time       `json:"created_at"`
}

type Case struct {
	ID        string    `json:"id"`
	CaseType  string    `json:"case_type"`
	EntityID  string    `json:"entity_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HandoffPacket struct {
	ID          string          `json:"id"`
	RunID       string          `json:"run_id"`
	CaseID      *string         `json:"case_id,omitempty"`
	HandoffType string          `json:"handoff_type"`
	FromPhase   int             `json:"from_phase"`
	ToPhase     int             `json:"to_phase"`
	PacketJSON  json.RawMessage `json:"packet_json"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}

type PhaseArtifact struct {
	ID          string          `json:"id"`
	CaseID      string          `json:"case_id"`
	Phase       int             `json:"phase"`
	Artifact    string          `json:"artifact_type"`
	ContentMD   *string         `json:"content_md,omitempty"`
	ContentJSON json.RawMessage `json:"content_json,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Decision struct {
	ID           string          `json:"id"`
	CaseID       string          `json:"case_id"`
	DecisionDate time.Time       `json:"decision_date"`
	OverallScore int             `json:"overall_score"`
	FinalLabel   string          `json:"final_label"`
	Constraints  json.RawMessage `json:"constraints_json"`
	JudgeResults json.RawMessage `json:"judge_results_json"`
	DecisionMD   string          `json:"decision_md"`
	CreatedAt    time.Time       `json:"created_at"`
}

type MonitoringPlan struct {
	ID         string          `json:"id"`
	CaseID     string          `json:"case_id"`
	DecisionID *string         `json:"decision_id,omitempty"`
	Status     string          `json:"status"`
	PlanJSON   json.RawMessage `json:"plan_json"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Alert struct {
	ID               string          `json:"id"`
	MonitoringPlanID string          `json:"monitoring_plan_id"`
	Severity         string          `json:"severity"`
	Type             string          `json:"type"`
	Message          string          `json:"message"`
	RefsJSON         json.RawMessage `json:"refs_json"`
	CreatedAt        time.Time       `json:"created_at"`
	AcknowledgedAt   *time.Time      `json:"acknowledged_at,omitempty"`
}

type Phase1RunEvent struct {
	RunID      string          `json:"run_id"`
	Seq        int             `json:"seq"`
	EventType  string          `json:"event_type"`
	Source     string          `json:"source"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload_json"`
	CreatedAt  time.Time       `json:"created_at"`
}
