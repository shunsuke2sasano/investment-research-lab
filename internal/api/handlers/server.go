package handlers

import "context"

type Server struct {
	store Store
}

func NewServer(store Store) *Server {
	return &Server{store: store}
}

type Store interface {
	CreateUniverseItem(ctx Context, item UniverseItemInput) (string, error)
	ListUniverseItems(ctx Context, f UniverseFilterInput) ([]UniverseItemOutput, *string, error)
	UpdateUniverseItem(ctx Context, id string, u UniverseUpdateInput) error

	CreateRun(ctx Context, mode string, configJSON []byte) (string, error)
	GetRun(ctx Context, id string) (RunOutput, error)
	ListEventsByRun(ctx Context, runID string, limit int, cursor string) ([]EventOutput, *string, error)
	AppendEventToRun(ctx Context, runID string, input RunEventInput) (int, error)
	ListPhase1RunEventsByRunID(ctx Context, runID string, limit int, cursor string) ([]Phase1RunEvent, *string, error)
	GetAnomalySummaryByRun(ctx Context, runID string) (AnomalySummaryOutput, error)
	GetTriggerDecisionByRun(ctx Context, runID string) (TriggerDecisionOutput, error)
	ListHandoffsByRun(ctx Context, runID string) ([]HandoffOutput, error)
	CreatePhase2Run(ctx Context, packet map[string]any) (string, error)
	UpdatePhase2RunPacket(ctx Context, runID string, packet map[string]any) error

	CreateHandoff(ctx Context, input HandoffInput) (string, error)
	GetHandoff(ctx Context, id string) (HandoffOutput, error)
	AttachCaseToHandoff(ctx Context, handoffID string, caseInput CaseInput) (string, error)

	CreateCase(ctx Context, input CaseInput) (string, error)
	ListCases(ctx Context, f CaseFilterInput) ([]CaseOutput, *string, error)
	GetCaseDetail(ctx Context, id string) (CaseDetailOutput, error)

	CreateArtifact(ctx Context, input ArtifactInput) (string, error)
	ListArtifacts(ctx Context, caseID string, f ArtifactFilterInput) ([]ArtifactOutput, error)

	CreateDecision(ctx Context, input DecisionInput) (string, error)
	ListDecisionsByCase(ctx Context, caseID string) ([]DecisionOutput, error)
	GetDecision(ctx Context, id string) (DecisionOutput, error)

	CreateMonitoringPlan(ctx Context, input MonitoringPlanInput) (string, error)
	GetMonitoringPlan(ctx Context, id string) (MonitoringPlanOutput, error)
	ListMonitoringPlansByCase(ctx Context, caseID string, limit int, cursor string) ([]MonitoringPlanOutput, *string, error)
	CreateAlert(ctx Context, input AlertInput) (string, error)
	ListAlertsByPlan(ctx Context, planID string, limit int, cursor string) ([]AlertOutput, *string, error)
	AckAlert(ctx Context, id string) error
}

type Context = context.Context
