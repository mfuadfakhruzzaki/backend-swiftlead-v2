package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/swiftlead/backend-swiftlet/internal/models"
)

var (
	ErrHarvestNotFound = errors.New("harvest not found")
)

// HarvestRepository interface
type HarvestRepository interface {
	Create(ctx context.Context, harvest *models.Harvest) error
	GetByID(ctx context.Context, id string) (*models.Harvest, error)
	Update(ctx context.Context, harvest *models.Harvest) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, rbwID string, limit, offset int) ([]*models.Harvest, int, error)
	GetLastHarvest(ctx context.Context, rbwID string, floorNo int) (*models.Harvest, error)
	GetStats(ctx context.Context, rbwID string) (*models.HarvestStats, error)
}

type harvestRepository struct {
	db *sql.DB
}

func NewHarvestRepository(db *sql.DB) HarvestRepository {
	return &harvestRepository{db: db}
}

func (r *harvestRepository) Create(ctx context.Context, harvest *models.Harvest) error {
	query := `
		INSERT INTO harvests (rbw_id, node_id, floor_no, harvested_at, nests_count, weight_kg, grade, notes, created_by, cycle_days)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		harvest.RBWID, harvest.NodeID, harvest.FloorNo, harvest.HarvestedAt,
		harvest.NestsCount, harvest.WeightKg, harvest.Grade, harvest.Notes,
		harvest.CreatedBy, harvest.CycleDays,
	).Scan(&harvest.ID, &harvest.CreatedAt, &harvest.UpdatedAt)
}

func (r *harvestRepository) GetByID(ctx context.Context, id string) (*models.Harvest, error) {
	query := `
		SELECT id, rbw_id, node_id, floor_no, harvested_at, nests_count, weight_kg,
		       grade, notes, created_by, cycle_days, created_at, updated_at
		FROM harvests WHERE id = $1
	`
	harvest := &models.Harvest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&harvest.ID, &harvest.RBWID, &harvest.NodeID, &harvest.FloorNo,
		&harvest.HarvestedAt, &harvest.NestsCount, &harvest.WeightKg,
		&harvest.Grade, &harvest.Notes, &harvest.CreatedBy, &harvest.CycleDays,
		&harvest.CreatedAt, &harvest.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrHarvestNotFound
	}
	return harvest, err
}

func (r *harvestRepository) Update(ctx context.Context, harvest *models.Harvest) error {
	query := `
		UPDATE harvests 
		SET floor_no = $1, nests_count = $2, weight_kg = $3, grade = $4, notes = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		harvest.FloorNo, harvest.NestsCount, harvest.WeightKg,
		harvest.Grade, harvest.Notes, harvest.ID,
	).Scan(&harvest.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrHarvestNotFound
	}
	return err
}

func (r *harvestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM harvests WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrHarvestNotFound
	}
	return nil
}

func (r *harvestRepository) List(ctx context.Context, rbwID string, limit, offset int) ([]*models.Harvest, int, error) {
	var harvests []*models.Harvest
	var total int

	countQuery := `SELECT COUNT(*) FROM harvests WHERE ($1 = '' OR rbw_id = $1)`
	if err := r.db.QueryRowContext(ctx, countQuery, rbwID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, rbw_id, node_id, floor_no, harvested_at, nests_count, weight_kg,
		       grade, notes, created_by, cycle_days, created_at, updated_at
		FROM harvests 
		WHERE ($1 = '' OR rbw_id = $1)
		ORDER BY harvested_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, rbwID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		harvest := &models.Harvest{}
		if err := rows.Scan(
			&harvest.ID, &harvest.RBWID, &harvest.NodeID, &harvest.FloorNo,
			&harvest.HarvestedAt, &harvest.NestsCount, &harvest.WeightKg,
			&harvest.Grade, &harvest.Notes, &harvest.CreatedBy, &harvest.CycleDays,
			&harvest.CreatedAt, &harvest.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		harvests = append(harvests, harvest)
	}

	return harvests, total, rows.Err()
}

func (r *harvestRepository) GetLastHarvest(ctx context.Context, rbwID string, floorNo int) (*models.Harvest, error) {
	query := `
		SELECT id, rbw_id, node_id, floor_no, harvested_at, nests_count, weight_kg,
		       grade, notes, created_by, cycle_days, created_at, updated_at
		FROM harvests 
		WHERE rbw_id = $1 AND floor_no = $2
		ORDER BY harvested_at DESC
		LIMIT 1
	`
	harvest := &models.Harvest{}
	err := r.db.QueryRowContext(ctx, query, rbwID, floorNo).Scan(
		&harvest.ID, &harvest.RBWID, &harvest.NodeID, &harvest.FloorNo,
		&harvest.HarvestedAt, &harvest.NestsCount, &harvest.WeightKg,
		&harvest.Grade, &harvest.Notes, &harvest.CreatedBy, &harvest.CycleDays,
		&harvest.CreatedAt, &harvest.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return harvest, err
}

func (r *harvestRepository) GetStats(ctx context.Context, rbwID string) (*models.HarvestStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_harvests,
			COALESCE(SUM(nests_count), 0) as total_nests,
			COALESCE(SUM(weight_kg), 0) as total_weight_kg,
			COALESCE(AVG(nests_count), 0) as avg_nests,
			COALESCE(AVG(weight_kg), 0) as avg_weight_kg,
			COALESCE(AVG(cycle_days), 0) as avg_cycle_days
		FROM harvests 
		WHERE ($1 = '' OR rbw_id = $1)
	`
	stats := &models.HarvestStats{}
	err := r.db.QueryRowContext(ctx, query, rbwID).Scan(
		&stats.TotalHarvests, &stats.TotalNests, &stats.TotalWeightKg,
		&stats.AvgNestsPerHarvest, &stats.AvgWeightKg, &stats.AvgCycleDays,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
