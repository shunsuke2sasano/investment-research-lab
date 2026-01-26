package queries

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"investment_committee/internal/db/models"
)

type UniverseFilter struct {
	Active     *bool
	EntityType *string
	Limit      int
	Cursor     *time.Time
}

type UniverseUpdate struct {
	Priority *int
	IsActive *bool
	Keywords []string
}

func (r *Repository) CreateUniverseItem(ctx context.Context, item models.UniverseItem) (string, error) {
	keywords, err := marshalJSON(jsonRawFromSlice(item.Keywords))
	if err != nil {
		return "", err
	}
	var id string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO universe_items (entity_type, entity_id, name, keywords, priority, is_active)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, item.EntityType, item.EntityID, item.Name, keywords, item.Priority, item.IsActive).Scan(&id)
	return id, err
}

func (r *Repository) ListUniverseItems(ctx context.Context, f UniverseFilter) ([]models.UniverseItem, *time.Time, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{}
	where := "WHERE 1=1"
	if f.Active != nil {
		args = append(args, *f.Active)
		where += " AND is_active = $" + itoa(len(args))
	}
	if f.EntityType != nil {
		args = append(args, *f.EntityType)
		where += " AND entity_type = $" + itoa(len(args))
	}
	if f.Cursor != nil {
		args = append(args, *f.Cursor)
		where += " AND created_at < $" + itoa(len(args))
	}
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, entity_type, entity_id, name, keywords, priority, is_active, created_at, updated_at
		FROM universe_items
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.UniverseItem
	var lastCreated *time.Time
	for rows.Next() {
		var item models.UniverseItem
		var keywords sql.NullString
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Name, &keywords, &item.Priority, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, nil, err
		}
		if keywords.Valid {
			item.Keywords = json.RawMessage(keywords.String)
		}
		items = append(items, item)
		t := item.CreatedAt
		lastCreated = &t
	}
	return items, lastCreated, rows.Err()
}

func (r *Repository) UpdateUniverseItem(ctx context.Context, id string, u UniverseUpdate) error {
	set := "SET updated_at = now()"
	args := []any{}
	if u.Priority != nil {
		args = append(args, *u.Priority)
		set += ", priority = $" + itoa(len(args))
	}
	if u.IsActive != nil {
		args = append(args, *u.IsActive)
		set += ", is_active = $" + itoa(len(args))
	}
	if u.Keywords != nil {
		raw, err := marshalJSON(u.Keywords)
		if err != nil {
			return err
		}
		args = append(args, raw)
		set += ", keywords = $" + itoa(len(args))
	}
	args = append(args, id)
	res, err := r.db.ExecContext(ctx, `
		UPDATE universe_items
		`+set+`
		WHERE id = $`+itoa(len(args)), args...)
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

func jsonRawFromSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	_ = json.Unmarshal(raw, &out)
	return out
}
