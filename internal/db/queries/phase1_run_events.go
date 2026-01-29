package queries

import (
	"context"
	"time"

	"investment_committee/internal/db/models"
)

func (r *Repository) CreatePhase1RunEvent(ctx context.Context, e models.Phase1RunEvent) (int, error) {
	var seq int
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1::text))`, e.RunID); err != nil {
		return 0, err
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO phase1_run_events (run_id, seq, event_type, source, occurred_at, payload_json)
		SELECT $1, COALESCE(MAX(seq), 0) + 1, $2, $3, $4, $5
		FROM phase1_run_events
		WHERE run_id = $1
		RETURNING seq
	`, e.RunID, e.EventType, e.Source, e.OccurredAt, e.Payload).Scan(&seq)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return seq, nil
}

func (r *Repository) ListPhase1RunEventsByRunID(ctx context.Context, runID string, limit int, cursor *time.Time) ([]models.Phase1RunEvent, *time.Time, error) {
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
		SELECT run_id, seq, event_type, source, occurred_at, payload_json, created_at
		FROM phase1_run_events
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.Phase1RunEvent
	var last *time.Time
	for rows.Next() {
		var e models.Phase1RunEvent
		if err := rows.Scan(&e.RunID, &e.Seq, &e.EventType, &e.Source, &e.OccurredAt, &e.Payload, &e.CreatedAt); err != nil {
			return nil, nil, err
		}
		items = append(items, e)
		t := e.CreatedAt
		last = &t
	}
	return items, last, rows.Err()
}
