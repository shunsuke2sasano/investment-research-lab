package phase1

import (
	"context"
	"encoding/json"
	"time"

	"investment_committee/internal/db/models"
	"investment_committee/internal/db/queries"
	"investment_committee/internal/domain"
)

func FinalizeRun(ctx context.Context, repo *queries.Repository, runID string) error {
	emptySummary := json.RawMessage(`{}`)
	if err := repo.CreateAnomalySummary(ctx, runID, emptySummary); err != nil {
		return err
	}
	decision := json.RawMessage(`{"should_handoff":false,"type":"none","reasons":[],"candidates":[]}`)
	if err := repo.CreateTriggerDecision(ctx, runID, decision); err != nil {
		return err
	}
	finalEvent := models.Phase1RunEvent{
		RunID:      runID,
		EventType:  domain.Phase1EventRunFinalized,
		Source:     "system",
		OccurredAt: time.Now().UTC(),
		Payload:    json.RawMessage(`{"status":"success"}`),
	}
	if _, err := repo.CreatePhase1RunEvent(ctx, finalEvent); err != nil {
		return err
	}
	return repo.UpdateRunStatus(ctx, runID, "success", nil)
}
