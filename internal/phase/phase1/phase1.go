package phase1

import (
	"context"
	"encoding/json"

	"investment_committee/internal/db/queries"
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
	return repo.UpdateRunStatus(ctx, runID, "success", nil)
}
