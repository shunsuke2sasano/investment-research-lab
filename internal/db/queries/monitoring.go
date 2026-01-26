package queries

import (
	"context"
	"database/sql"
	"time"

	"investment_committee/internal/db/models"
)

func (r *Repository) CreateMonitoringPlan(ctx context.Context, m models.MonitoringPlan) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO monitoring_plans (case_id, decision_id, status, plan_json)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, m.CaseID, m.DecisionID, m.Status, m.PlanJSON).Scan(&id)
	return id, err
}

func (r *Repository) GetMonitoringPlan(ctx context.Context, id string) (models.MonitoringPlan, error) {
	var m models.MonitoringPlan
	var decisionID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, case_id, decision_id, status, plan_json, created_at, updated_at
		FROM monitoring_plans
		WHERE id = $1
	`, id).Scan(&m.ID, &m.CaseID, &decisionID, &m.Status, &m.PlanJSON, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return m, ErrNotFound
	}
	if err != nil {
		return m, err
	}
	if decisionID.Valid {
		v := decisionID.String
		m.DecisionID = &v
	}
	return m, nil
}

func (r *Repository) ListMonitoringPlansByCase(ctx context.Context, caseID string, limit int, cursor *time.Time) ([]models.MonitoringPlan, *time.Time, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{caseID}
	where := "WHERE case_id = $1"
	if cursor != nil {
		args = append(args, *cursor)
		where += " AND created_at < $" + itoa(len(args))
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, case_id, decision_id, status, plan_json, created_at, updated_at
		FROM monitoring_plans
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.MonitoringPlan
	var last *time.Time
	for rows.Next() {
		var m models.MonitoringPlan
		var decisionID sql.NullString
		if err := rows.Scan(&m.ID, &m.CaseID, &decisionID, &m.Status, &m.PlanJSON, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, nil, err
		}
		if decisionID.Valid {
			v := decisionID.String
			m.DecisionID = &v
		}
		items = append(items, m)
		t := m.CreatedAt
		last = &t
	}
	return items, last, rows.Err()
}

func (r *Repository) CreateAlert(ctx context.Context, a models.Alert) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO alerts (monitoring_plan_id, severity, type, message, refs_json)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`, a.MonitoringPlanID, a.Severity, a.Type, a.Message, a.RefsJSON).Scan(&id)
	return id, err
}

func (r *Repository) ListAlertsByPlan(ctx context.Context, planID string, limit int, cursor *time.Time) ([]models.Alert, *time.Time, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{planID}
	where := "WHERE monitoring_plan_id = $1"
	if cursor != nil {
		args = append(args, *cursor)
		where += " AND created_at < $" + itoa(len(args))
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, monitoring_plan_id, severity, type, message, refs_json, created_at, acknowledged_at
		FROM alerts
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.Alert
	var last *time.Time
	for rows.Next() {
		var a models.Alert
		var ack sql.NullTime
		if err := rows.Scan(&a.ID, &a.MonitoringPlanID, &a.Severity, &a.Type, &a.Message, &a.RefsJSON, &a.CreatedAt, &ack); err != nil {
			return nil, nil, err
		}
		if ack.Valid {
			t := ack.Time
			a.AcknowledgedAt = &t
		}
		items = append(items, a)
		t := a.CreatedAt
		last = &t
	}
	return items, last, rows.Err()
}

func (r *Repository) AckAlert(ctx context.Context, alertID string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE alerts
		SET acknowledged_at = now()
		WHERE id = $1
	`, alertID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}
