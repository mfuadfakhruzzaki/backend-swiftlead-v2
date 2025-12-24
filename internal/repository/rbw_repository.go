package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/swiftlead/backend-swiftlet/internal/models"
)

var (
	ErrRBWNotFound      = errors.New("RBW not found")
	ErrRBWAlreadyExists = errors.New("RBW with this code already exists")
)

// RBWRepository interface
type RBWRepository interface {
	Create(ctx context.Context, rbw *models.RBW) error
	GetByID(ctx context.Context, id string) (*models.RBW, error)
	Update(ctx context.Context, rbw *models.RBW) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, ownerID string, limit, offset int) ([]*models.RBW, int, error)
}

type rbwRepository struct {
	db *sql.DB
}

func NewRBWRepository(db *sql.DB) RBWRepository {
	return &rbwRepository{db: db}
}

func (r *rbwRepository) Create(ctx context.Context, rbw *models.RBW) error {
	query := `
		INSERT INTO rbw (owner_id, code, name, address, latitude, longitude, total_floors, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		rbw.OwnerID, rbw.Code, rbw.Name, rbw.Address,
		rbw.Latitude, rbw.Longitude, rbw.TotalFloors, rbw.Description,
	).Scan(&rbw.ID, &rbw.CreatedAt, &rbw.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrRBWAlreadyExists
		}
		return err
	}
	return nil
}

func (r *rbwRepository) GetByID(ctx context.Context, id string) (*models.RBW, error) {
	query := `
		SELECT id, owner_id, code, name, address, latitude, longitude, 
		       total_floors, description, photo_url, created_at, updated_at
		FROM rbw WHERE id = $1
	`
	rbw := &models.RBW{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rbw.ID, &rbw.OwnerID, &rbw.Code, &rbw.Name, &rbw.Address,
		&rbw.Latitude, &rbw.Longitude, &rbw.TotalFloors, &rbw.Description,
		&rbw.PhotoURL, &rbw.CreatedAt, &rbw.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRBWNotFound
	}
	if err != nil {
		return nil, err
	}
	return rbw, nil
}

func (r *rbwRepository) Update(ctx context.Context, rbw *models.RBW) error {
	query := `
		UPDATE rbw 
		SET name = $1, address = $2, latitude = $3, longitude = $4, 
		    total_floors = $5, description = $6, photo_url = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		rbw.Name, rbw.Address, rbw.Latitude, rbw.Longitude,
		rbw.TotalFloors, rbw.Description, rbw.PhotoURL, rbw.ID,
	).Scan(&rbw.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrRBWNotFound
	}
	return err
}

func (r *rbwRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM rbw WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrRBWNotFound
	}
	return nil
}

func (r *rbwRepository) List(ctx context.Context, ownerID string, limit, offset int) ([]*models.RBW, int, error) {
	var rbws []*models.RBW
	var total int

	countQuery := `SELECT COUNT(*) FROM rbw WHERE ($1 = '' OR owner_id::text = $1)`
	if err := r.db.QueryRowContext(ctx, countQuery, ownerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, owner_id, code, name, address, latitude, longitude, 
		       total_floors, description, photo_url, created_at, updated_at
		FROM rbw 
		WHERE ($1 = '' OR owner_id::text = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		rbw := &models.RBW{}
		if err := rows.Scan(
			&rbw.ID, &rbw.OwnerID, &rbw.Code, &rbw.Name, &rbw.Address,
			&rbw.Latitude, &rbw.Longitude, &rbw.TotalFloors, &rbw.Description,
			&rbw.PhotoURL, &rbw.CreatedAt, &rbw.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		rbws = append(rbws, rbw)
	}

	return rbws, total, rows.Err()
}
