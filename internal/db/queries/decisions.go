package queries

import (
	"context"
	"database/sql"

	"investment_committee/internal/db/models"
)

func (r *Repository) CreateDecision(ctx context.Context, d models.Decision) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO decisions (case_id, overall_score, final_label, constraints_json, judge_results_json, decision_md)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, d.CaseID, d.OverallScore, d.FinalLabel, d.Constraints, d.JudgeResults, d.DecisionMD).Scan(&id)
	return id, err
}

func (r *Repository) ListDecisionsByCase(ctx context.Context, caseID string) ([]models.Decision, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, case_id, decision_date, overall_score, final_label, constraints_json, judge_results_json, decision_md, created_at
		FROM decisions
		WHERE case_id = $1
		ORDER BY decision_date DESC
	`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Decision
	for rows.Next() {
		var d models.Decision
		if err := rows.Scan(&d.ID, &d.CaseID, &d.DecisionDate, &d.OverallScore, &d.FinalLabel, &d.Constraints, &d.JudgeResults, &d.DecisionMD, &d.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *Repository) GetDecision(ctx context.Context, id string) (models.Decision, error) {
	var d models.Decision
	err := r.db.QueryRowContext(ctx, `
		SELECT id, case_id, decision_date, overall_score, final_label, constraints_json, judge_results_json, decision_md, created_at
		FROM decisions
		WHERE id = $1
	`, id).Scan(&d.ID, &d.CaseID, &d.DecisionDate, &d.OverallScore, &d.FinalLabel, &d.Constraints, &d.JudgeResults, &d.DecisionMD, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return d, ErrNotFound
	}
	return d, err
}
