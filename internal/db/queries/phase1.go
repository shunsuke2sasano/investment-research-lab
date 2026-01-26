package queries

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"investment_committee/internal/db/models"
)

func (r *Repository) CreateRun(ctx context.Context, mode string, config json.RawMessage) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO runs (phase, mode, status, config_json)
		VALUES (1, $1, 'running', $2)
		RETURNING id
	`, mode, config).Scan(&id)
	return id, err
}

func (r *Repository) UpdateRunStatus(ctx context.Context, runID string, status string, errMsg *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE runs
		SET status = $1, finished_at = now(), error = $2
		WHERE id = $3
	`, status, errMsg, runID)
	return err
}

func (r *Repository) GetRun(ctx context.Context, runID string) (models.Run, error) {
	var run models.Run
	var finishedAt sql.NullTime
	var errMsg sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phase, mode, status, config_json, started_at, finished_at, error
		FROM runs
		WHERE id = $1
	`, runID).Scan(&run.ID, &run.Phase, &run.Mode, &run.Status, &run.ConfigJSON, &run.StartedAt, &finishedAt, &errMsg)
	if err == sql.ErrNoRows {
		return run, ErrNotFound
	}
	if err != nil {
		return run, err
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		run.FinishedAt = &t
	}
	if errMsg.Valid {
		s := errMsg.String
		run.Error = &s
	}
	return run, nil
}

func (r *Repository) ListEventsByRun(ctx context.Context, runID string, limit int, cursor *time.Time) ([]models.Event, *time.Time, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{runID}
	where := "WHERE run_id = $1"
	if cursor != nil {
		args = append(args, *cursor)
		where += " AND created_at < $" + itoa(len(args))
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `
		SELECT event_id, run_id, observed_at, entity_type, entity_id, category, title,
		       facts_json, impact_json, sources_json, confidence, dedupe_key, tags_json, created_at
		FROM events
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.Event
	var last *time.Time
	for rows.Next() {
		var e models.Event
		var impact sql.NullString
		var tags sql.NullString
		if err := rows.Scan(&e.EventID, &e.RunID, &e.ObservedAt, &e.EntityType, &e.EntityID, &e.Category, &e.Title,
			&e.FactsJSON, &impact, &e.Sources, &e.Confidence, &e.DedupeKey, &tags, &e.CreatedAt); err != nil {
			return nil, nil, err
		}
		if impact.Valid {
			e.ImpactJSON = json.RawMessage(impact.String)
		}
		if tags.Valid {
			e.TagsJSON = json.RawMessage(tags.String)
		}
		items = append(items, e)
		t := e.CreatedAt
		last = &t
	}
	return items, last, rows.Err()
}

func (r *Repository) GetAnomalySummaryByRun(ctx context.Context, runID string) (models.AnomalySummary, error) {
	var a models.AnomalySummary
	err := r.db.QueryRowContext(ctx, `
		SELECT run_id, summary_json, created_at
		FROM anomaly_summaries
		WHERE run_id = $1
	`, runID).Scan(&a.RunID, &a.Summary, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return a, ErrNotFound
	}
	return a, err
}

func (r *Repository) GetTriggerDecisionByRun(ctx context.Context, runID string) (models.TriggerDecision, error) {
	var t models.TriggerDecision
	err := r.db.QueryRowContext(ctx, `
		SELECT run_id, decision_json, created_at
		FROM trigger_decisions
		WHERE run_id = $1
	`, runID).Scan(&t.RunID, &t.Decision, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return t, ErrNotFound
	}
	return t, err
}

func (r *Repository) ListHandoffsByRun(ctx context.Context, runID string) ([]models.HandoffPacket, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, run_id, case_id, handoff_type, from_phase, to_phase, packet_json, status, created_at
		FROM handoff_packets
		WHERE run_id = $1
		ORDER BY created_at DESC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.HandoffPacket
	for rows.Next() {
		var h models.HandoffPacket
		var caseID sql.NullString
		if err := rows.Scan(&h.ID, &h.RunID, &caseID, &h.HandoffType, &h.FromPhase, &h.ToPhase, &h.PacketJSON, &h.Status, &h.CreatedAt); err != nil {
			return nil, err
		}
		if caseID.Valid {
			id := caseID.String
			h.CaseID = &id
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

func (r *Repository) CreateAnomalySummary(ctx context.Context, runID string, summary json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO anomaly_summaries (run_id, summary_json)
		VALUES ($1, $2)
		ON CONFLICT (run_id) DO UPDATE SET summary_json = EXCLUDED.summary_json, created_at = now()
	`, runID, summary)
	return err
}

func (r *Repository) CreateTriggerDecision(ctx context.Context, runID string, decision json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO trigger_decisions (run_id, decision_json)
		VALUES ($1, $2)
		ON CONFLICT (run_id) DO UPDATE SET decision_json = EXCLUDED.decision_json, created_at = now()
	`, runID, decision)
	return err
}

func (r *Repository) CreateHandoffPacket(ctx context.Context, h models.HandoffPacket) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO handoff_packets (run_id, case_id, handoff_type, from_phase, to_phase, packet_json, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`, h.RunID, h.CaseID, h.HandoffType, h.FromPhase, h.ToPhase, h.PacketJSON, h.Status).Scan(&id)
	return id, err
}
