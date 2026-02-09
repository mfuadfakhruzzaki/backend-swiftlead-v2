package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/swiftlead/backend-swiftlet/internal/models"
)

var (
	ErrAlertNotFound = errors.New("alert not found")
)

// AlertRepository interface
type AlertRepository interface {
	Create(ctx context.Context, alert *models.Alert) error
	GetByID(ctx context.Context, id string) (*models.Alert, error)
	MarkAsRead(ctx context.Context, id string) error
	Resolve(ctx context.Context, id, resolvedBy string) error
	List(ctx context.Context, rbwID string, isRead, resolved *bool, limit, offset int) ([]*models.Alert, int, error)
	ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Alert, int, error)
}

type alertRepository struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) AlertRepository {
	return &alertRepository{db: db}
}

func (r *alertRepository) Create(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (rbw_id, node_id, sensor_id, alert_type, severity, message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_read, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		alert.RBWID, alert.NodeID, alert.SensorID,
		alert.AlertType, alert.Severity, alert.Message,
	).Scan(&alert.ID, &alert.IsRead, &alert.CreatedAt)
}

func (r *alertRepository) GetByID(ctx context.Context, id string) (*models.Alert, error) {
	query := `
		SELECT id, rbw_id, node_id, sensor_id, alert_type, severity, message,
		       is_read, resolved_at, resolved_by, created_at
		FROM alerts WHERE id = $1
	`
	alert := &models.Alert{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.RBWID, &alert.NodeID, &alert.SensorID,
		&alert.AlertType, &alert.Severity, &alert.Message,
		&alert.IsRead, &alert.ResolvedAt, &alert.ResolvedBy, &alert.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrAlertNotFound
	}
	return alert, err
}

func (r *alertRepository) MarkAsRead(ctx context.Context, id string) error {
	query := `UPDATE alerts SET is_read = true WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAlertNotFound
	}
	return nil
}

func (r *alertRepository) Resolve(ctx context.Context, id, resolvedBy string) error {
	query := `UPDATE alerts SET resolved_at = NOW(), resolved_by = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, resolvedBy, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAlertNotFound
	}
	return nil
}

func (r *alertRepository) List(ctx context.Context, rbwID string, isRead, resolved *bool, limit, offset int) ([]*models.Alert, int, error) {
	var alerts []*models.Alert
	var total int

	// Build query dynamically based on filters
	baseWhere := "WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if rbwID != "" {
		baseWhere += fmt.Sprintf(" AND rbw_id = $%d", argNum)
		args = append(args, rbwID)
		argNum++
	}

	if isRead != nil {
		baseWhere += fmt.Sprintf(" AND is_read = $%d", argNum)
		args = append(args, *isRead)
		argNum++
	}

	if resolved != nil {
		if *resolved {
			baseWhere += " AND resolved_at IS NOT NULL"
		} else {
			baseWhere += " AND resolved_at IS NULL"
		}
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM alerts " + baseWhere
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data query with pagination
	query := fmt.Sprintf(`
		SELECT id, rbw_id, node_id, sensor_id, alert_type, severity, message,
		       is_read, resolved_at, resolved_by, created_at
		FROM alerts %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseWhere, argNum, argNum+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		alert := &models.Alert{}
		if err := rows.Scan(
			&alert.ID, &alert.RBWID, &alert.NodeID, &alert.SensorID,
			&alert.AlertType, &alert.Severity, &alert.Message,
			&alert.IsRead, &alert.ResolvedAt, &alert.ResolvedBy, &alert.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, total, rows.Err()
}

func (r *alertRepository) ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Alert, int, error) {
	return r.List(ctx, rbwID, nil, nil, limit, offset)
}
