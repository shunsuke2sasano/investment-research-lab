package handlers

import "time"

type UniverseItemInput struct {
	EntityType string   `json:"entity_type"`
	EntityID   string   `json:"entity_id"`
	Name       string   `json:"name"`
	Keywords   []string `json:"keywords,omitempty"`
	Priority   int      `json:"priority"`
	IsActive   bool     `json:"is_active"`
}

type UniverseItemOutput struct {
	ID         string   `json:"id"`
	EntityType string   `json:"entity_type"`
	EntityID   string   `json:"entity_id"`
	Name       string   `json:"name"`
	Keywords   []string `json:"keywords,omitempty"`
	Priority   int      `json:"priority"`
	IsActive   bool     `json:"is_active"`
}

type UniverseFilterInput struct {
	Active     *bool
	EntityType *string
	Limit      int
	Cursor     string
}

type UniverseUpdateInput struct {
	Priority *int     `json:"priority,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

type RunInput struct {
	Mode   string         `json:"mode"`
	Config map[string]any `json:"config"`
}

type RunOutput struct {
	ID         string         `json:"id"`
	Status     string         `json:"status"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	Error      *string        `json:"error,omitempty"`
	Config     map[string]any `json:"config,omitempty"`
	Events     []Phase1RunEvent `json:"events"`
}

type Phase2RunInput struct {
	Packet map[string]any `json:"packet"`
}

type Phase2RunOutput struct {
	RunID  string         `json:"run_id"`
	Packet map[string]any `json:"packet"`
}

type Phase3RunInput struct {
	Packet map[string]any `json:"packet"`
}

type Phase3RunOutput struct {
	RunID  string         `json:"run_id"`
	Packet map[string]any `json:"packet"`
}

type Phase4RunInput struct {
	Packet map[string]any `json:"packet"`
}

type Phase4RunOutput struct {
	RunID  string         `json:"run_id"`
	Packet map[string]any `json:"packet"`
}

type EventOutput struct {
	EventID    string    `json:"event_id"`
	Category   string    `json:"category"`
	ObservedAt time.Time `json:"observed_at"`
	Title      string    `json:"title"`
	Facts      any       `json:"facts_json"`
	Impact     any       `json:"impact_json,omitempty"`
	Sources    any       `json:"sources_json"`
	Confidence float64   `json:"confidence"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
}

type AnomalySummaryOutput struct {
	RunID   string `json:"run_id"`
	Summary any    `json:"summary_json"`
}

type TriggerDecisionOutput struct {
	RunID    string `json:"run_id"`
	Decision any    `json:"decision_json"`
}

type RunEventInput struct {
	EventType  string         `json:"event_type"`
	Source     string         `json:"source,omitempty"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type Phase1RunEvent struct {
	RunID      string         `json:"run_id"`
	Seq        int            `json:"seq"`
	EventType  string         `json:"event_type"`
	Source     string         `json:"source"`
	OccurredAt time.Time      `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

type HandoffInput struct {
	RunID       string         `json:"run_id"`
	CaseID      *string        `json:"case_id,omitempty"`
	HandoffType string         `json:"handoff_type"`
	FromPhase   int            `json:"from_phase"`
	ToPhase     int            `json:"to_phase"`
	Packet      map[string]any `json:"packet"`
}

type HandoffOutput struct {
	ID          string         `json:"id"`
	RunID       string         `json:"run_id"`
	CaseID      *string        `json:"case_id,omitempty"`
	HandoffType string         `json:"handoff_type"`
	FromPhase   int            `json:"from_phase"`
	ToPhase     int            `json:"to_phase"`
	Packet      map[string]any `json:"packet"`
	Status      string         `json:"status"`
}

type CaseInput struct {
	CaseType string `json:"case_type"`
	EntityID string `json:"entity_id"`
	Title    string `json:"title"`
	Priority int    `json:"priority"`
}

type CaseOutput struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	CaseType string `json:"case_type"`
	EntityID string `json:"entity_id"`
	Priority int    `json:"priority"`
}

type CaseFilterInput struct {
	Status *string
	Limit  int
	Cursor string
}

type CaseDetailOutput struct {
	Case            CaseOutput            `json:"case"`
	Handoffs        []HandoffOutput       `json:"handoffs"`
	Artifacts       []ArtifactOutput      `json:"artifacts"`
	Decisions       []DecisionOutput      `json:"decisions"`
	MonitoringPlans []MonitoringPlanOutput `json:"monitoring_plans"`
}

type ArtifactInput struct {
	CaseID       string         `json:"case_id"`
	Phase        int            `json:"phase"`
	ArtifactType string         `json:"artifact_type"`
	ContentMD    *string        `json:"content_md,omitempty"`
	ContentJSON  map[string]any `json:"content_json,omitempty"`
}

type ArtifactFilterInput struct {
	Phase  *int
	Latest bool
}

type ArtifactOutput struct {
	ID           string         `json:"id"`
	Phase        int            `json:"phase"`
	ArtifactType string         `json:"artifact_type"`
	ContentMD    *string        `json:"content_md,omitempty"`
	ContentJSON  map[string]any `json:"content_json,omitempty"`
}

type DecisionInput struct {
	CaseID       string         `json:"case_id"`
	OverallScore int            `json:"overall_score"`
	FinalLabel   string         `json:"final_label"`
	Constraints  []string       `json:"constraints"`
	JudgeResults map[string]any `json:"judge_results"`
	DecisionMD   string         `json:"decision_md"`
}

type DecisionOutput struct {
	ID           string         `json:"id"`
	CaseID       string         `json:"case_id"`
	DecisionDate time.Time      `json:"decision_date"`
	OverallScore int            `json:"overall_score"`
	FinalLabel   string         `json:"final_label"`
	Constraints  []string       `json:"constraints"`
	JudgeResults map[string]any `json:"judge_results"`
	DecisionMD   string         `json:"decision_md"`
}

type MonitoringPlanInput struct {
	CaseID     string         `json:"case_id"`
	DecisionID *string        `json:"decision_id,omitempty"`
	Plan       map[string]any `json:"plan"`
}

type MonitoringPlanOutput struct {
	ID         string         `json:"id"`
	CaseID     string         `json:"case_id"`
	DecisionID *string        `json:"decision_id,omitempty"`
	Status     string         `json:"status"`
	Plan       map[string]any `json:"plan"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type AlertInput struct {
	MonitoringPlanID string         `json:"monitoring_plan_id"`
	Severity         string         `json:"severity"`
	Type             string         `json:"type"`
	Message          string         `json:"message"`
	Refs             map[string]any `json:"refs"`
}

type AlertOutput struct {
	ID        string         `json:"id"`
	Severity  string         `json:"severity"`
	Type      string         `json:"type"`
	Message   string         `json:"message"`
	Refs      map[string]any `json:"refs"`
	CreatedAt time.Time      `json:"created_at"`
	AckAt     *time.Time     `json:"acknowledged_at,omitempty"`
}
