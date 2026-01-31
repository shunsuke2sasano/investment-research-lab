package queries

import (
	"context"
	"encoding/json"
)

func (r *Repository) CreatePhase2Run(ctx context.Context, packet json.RawMessage) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO phase2_runs (input_packet)
		VALUES ($1)
		RETURNING id
	`, packet).Scan(&id)
	return id, err
}

func (r *Repository) UpdatePhase2RunPacket(ctx context.Context, runID string, packet json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE phase2_runs
		SET input_packet = $1
		WHERE id = $2
	`, packet, runID)
	return err
}
