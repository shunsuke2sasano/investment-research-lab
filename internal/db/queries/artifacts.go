package queries

import (
	"context"
	"database/sql"

	"investment_committee/internal/db/models"
)

type ArtifactFilter struct {
	Phase  *int
	Latest bool
}

func (r *Repository) CreateArtifact(ctx context.Context, a models.PhaseArtifact) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO phase_artifacts (case_id, phase, artifact_type, content_md, content_json)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`, a.CaseID, a.Phase, a.Artifact, a.ContentMD, a.ContentJSON).Scan(&id)
	return id, err
}

func (r *Repository) ListArtifacts(ctx context.Context, caseID string, f ArtifactFilter) ([]models.PhaseArtifact, error) {
	where := "WHERE case_id = $1"
	args := []any{caseID}
	if f.Phase != nil {
		args = append(args, *f.Phase)
		where += " AND phase = $" + itoa(len(args))
	}
	order := "ORDER BY created_at DESC"
	limit := ""
	if f.Latest {
		limit = " LIMIT 1"
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, case_id, phase, artifact_type, content_md, content_json, created_at
		FROM phase_artifacts
		`+where+`
		`+order+limit, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PhaseArtifact
	for rows.Next() {
		var a models.PhaseArtifact
		var contentMD sql.NullString
		var contentJSON sql.NullString
		if err := rows.Scan(&a.ID, &a.CaseID, &a.Phase, &a.Artifact, &contentMD, &contentJSON, &a.CreatedAt); err != nil {
			return nil, err
		}
		if contentMD.Valid {
			v := contentMD.String
			a.ContentMD = &v
		}
		if contentJSON.Valid {
			a.ContentJSON = []byte(contentJSON.String)
		}
		items = append(items, a)
	}
	return items, rows.Err()
}
