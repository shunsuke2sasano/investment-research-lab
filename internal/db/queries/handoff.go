package queries

import (
	"context"
	"database/sql"

	"investment_committee/internal/db/models"
)

func (r *Repository) CreateHandoff(ctx context.Context, h models.HandoffPacket) (string, error) {
	return r.CreateHandoffPacket(ctx, h)
}

func (r *Repository) GetHandoff(ctx context.Context, id string) (models.HandoffPacket, error) {
	var h models.HandoffPacket
	var caseID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, run_id, case_id, handoff_type, from_phase, to_phase, packet_json, status, created_at
		FROM handoff_packets
		WHERE id = $1
	`, id).Scan(&h.ID, &h.RunID, &caseID, &h.HandoffType, &h.FromPhase, &h.ToPhase, &h.PacketJSON, &h.Status, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return h, ErrNotFound
	}
	if err != nil {
		return h, err
	}
	if caseID.Valid {
		v := caseID.String
		h.CaseID = &v
	}
	return h, nil
}

func (r *Repository) AttachCaseToHandoff(ctx context.Context, handoffID string, caseID string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE handoff_packets
		SET case_id = $1
		WHERE id = $2
	`, caseID, handoffID)
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
