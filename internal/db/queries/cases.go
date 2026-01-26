package queries

import (
	"context"
	"database/sql"
	"time"

	"investment_committee/internal/db/models"
)

type CaseFilter struct {
	Status *string
	Limit  int
	Cursor *time.Time
}

type CaseDetail struct {
	Case            models.Case             `json:"case"`
	Handoffs        []models.HandoffPacket  `json:"handoffs"`
	Artifacts       []models.PhaseArtifact  `json:"artifacts"`
	Decisions       []models.Decision       `json:"decisions"`
	MonitoringPlans []models.MonitoringPlan `json:"monitoring_plans"`
}

func (r *Repository) CreateCase(ctx context.Context, c models.Case) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO cases (case_type, entity_id, title, priority)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, c.CaseType, c.EntityID, c.Title, c.Priority).Scan(&id)
	return id, err
}

func (r *Repository) ListCases(ctx context.Context, f CaseFilter) ([]models.Case, *time.Time, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{}
	where := "WHERE 1=1"
	if f.Status != nil {
		args = append(args, *f.Status)
		where += " AND status = $" + itoa(len(args))
	}
	if f.Cursor != nil {
		args = append(args, *f.Cursor)
		where += " AND created_at < $" + itoa(len(args))
	}
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, case_type, entity_id, title, status, priority, created_at, updated_at
		FROM cases
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.Case
	var last *time.Time
	for rows.Next() {
		var c models.Case
		if err := rows.Scan(&c.ID, &c.CaseType, &c.EntityID, &c.Title, &c.Status, &c.Priority, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, nil, err
		}
		items = append(items, c)
		t := c.CreatedAt
		last = &t
	}
	return items, last, rows.Err()
}

func (r *Repository) GetCaseDetail(ctx context.Context, id string) (CaseDetail, error) {
	var d CaseDetail
	var c models.Case
	err := r.db.QueryRowContext(ctx, `
		SELECT id, case_type, entity_id, title, status, priority, created_at, updated_at
		FROM cases
		WHERE id = $1
	`, id).Scan(&c.ID, &c.CaseType, &c.EntityID, &c.Title, &c.Status, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return d, ErrNotFound
	}
	if err != nil {
		return d, err
	}
	d.Case = c

	hRows, err := r.db.QueryContext(ctx, `
		SELECT id, run_id, case_id, handoff_type, from_phase, to_phase, packet_json, status, created_at
		FROM handoff_packets
		WHERE case_id = $1
		ORDER BY created_at DESC
	`, id)
	if err != nil {
		return d, err
	}
	for hRows.Next() {
		var h models.HandoffPacket
		var caseID sql.NullString
		if err := hRows.Scan(&h.ID, &h.RunID, &caseID, &h.HandoffType, &h.FromPhase, &h.ToPhase, &h.PacketJSON, &h.Status, &h.CreatedAt); err != nil {
			hRows.Close()
			return d, err
		}
		if caseID.Valid {
			v := caseID.String
			h.CaseID = &v
		}
		d.Handoffs = append(d.Handoffs, h)
	}
	hRows.Close()

	aRows, err := r.db.QueryContext(ctx, `
		SELECT id, case_id, phase, artifact_type, content_md, content_json, created_at
		FROM phase_artifacts
		WHERE case_id = $1
		ORDER BY created_at DESC
	`, id)
	if err != nil {
		return d, err
	}
	for aRows.Next() {
		var a models.PhaseArtifact
		var contentMD sql.NullString
		var contentJSON sql.NullString
		if err := aRows.Scan(&a.ID, &a.CaseID, &a.Phase, &a.Artifact, &contentMD, &contentJSON, &a.CreatedAt); err != nil {
			aRows.Close()
			return d, err
		}
		if contentMD.Valid {
			v := contentMD.String
			a.ContentMD = &v
		}
		if contentJSON.Valid {
			a.ContentJSON = []byte(contentJSON.String)
		}
		d.Artifacts = append(d.Artifacts, a)
	}
	aRows.Close()

	dRows, err := r.db.QueryContext(ctx, `
		SELECT id, case_id, decision_date, overall_score, final_label, constraints_json, judge_results_json, decision_md, created_at
		FROM decisions
		WHERE case_id = $1
		ORDER BY decision_date DESC
	`, id)
	if err != nil {
		return d, err
	}
	for dRows.Next() {
		var dec models.Decision
		if err := dRows.Scan(&dec.ID, &dec.CaseID, &dec.DecisionDate, &dec.OverallScore, &dec.FinalLabel, &dec.Constraints, &dec.JudgeResults, &dec.DecisionMD, &dec.CreatedAt); err != nil {
			dRows.Close()
			return d, err
		}
		d.Decisions = append(d.Decisions, dec)
	}
	dRows.Close()

	mRows, err := r.db.QueryContext(ctx, `
		SELECT id, case_id, decision_id, status, plan_json, created_at, updated_at
		FROM monitoring_plans
		WHERE case_id = $1
		ORDER BY created_at DESC
	`, id)
	if err != nil {
		return d, err
	}
	for mRows.Next() {
		var m models.MonitoringPlan
		var decisionID sql.NullString
		if err := mRows.Scan(&m.ID, &m.CaseID, &decisionID, &m.Status, &m.PlanJSON, &m.CreatedAt, &m.UpdatedAt); err != nil {
			mRows.Close()
			return d, err
		}
		if decisionID.Valid {
			v := decisionID.String
			m.DecisionID = &v
		}
		d.MonitoringPlans = append(d.MonitoringPlans, m)
	}
	mRows.Close()

	return d, nil
}
