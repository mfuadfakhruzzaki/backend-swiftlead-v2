package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/swiftlead/backend-swiftlet/internal/models"
)

var (
	ErrServiceRequestNotFound = errors.New("service request not found")
	ErrTransactionNotFound    = errors.New("transaction not found")
)

// ServiceRequestRepository interface
type ServiceRequestRepository interface {
	Create(ctx context.Context, sr *models.ServiceRequest) error
	GetByID(ctx context.Context, id string) (*models.ServiceRequest, error)
	Update(ctx context.Context, sr *models.ServiceRequest) error
	List(ctx context.Context, rbwID, requestBy, assignedTo, status string, limit, offset int) ([]*models.ServiceRequest, int, error)
}

type serviceRequestRepository struct {
	db *sql.DB
}

func NewServiceRequestRepository(db *sql.DB) ServiceRequestRepository {
	return &serviceRequestRepository{db: db}
}

func (r *serviceRequestRepository) Create(ctx context.Context, sr *models.ServiceRequest) error {
	query := `
		INSERT INTO service_requests (rbw_id, node_id, request_by, type, status, issue)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, request_date, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		sr.RBWID, sr.NodeID, sr.RequestBy, sr.Type, sr.Status, sr.Issue,
	).Scan(&sr.ID, &sr.RequestDate, &sr.CreatedAt, &sr.UpdatedAt)
}

func (r *serviceRequestRepository) GetByID(ctx context.Context, id string) (*models.ServiceRequest, error) {
	query := `
		SELECT id, rbw_id, node_id, request_by, assigned_to, approved_by, type, status,
		       request_date, schedule_date, uninstall_date, issue, resolution, notes,
		       created_at, updated_at
		FROM service_requests WHERE id = $1
	`
	sr := &models.ServiceRequest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sr.ID, &sr.RBWID, &sr.NodeID, &sr.RequestBy, &sr.AssignedTo, &sr.ApprovedBy,
		&sr.Type, &sr.Status, &sr.RequestDate, &sr.ScheduleDate, &sr.UninstallDate,
		&sr.Issue, &sr.Resolution, &sr.Notes, &sr.CreatedAt, &sr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrServiceRequestNotFound
	}
	return sr, err
}

func (r *serviceRequestRepository) Update(ctx context.Context, sr *models.ServiceRequest) error {
	query := `
		UPDATE service_requests 
		SET status = $1, assigned_to = $2, approved_by = $3, schedule_date = $4,
		    resolution = $5, notes = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		sr.Status, sr.AssignedTo, sr.ApprovedBy, sr.ScheduleDate,
		sr.Resolution, sr.Notes, sr.ID,
	).Scan(&sr.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrServiceRequestNotFound
	}
	return err
}

func (r *serviceRequestRepository) List(ctx context.Context, rbwID, requestBy, assignedTo, status string, limit, offset int) ([]*models.ServiceRequest, int, error) {
	var srs []*models.ServiceRequest
	var total int

	countQuery := `
		SELECT COUNT(*) FROM service_requests 
		WHERE ($1 = '' OR rbw_id = $1)
		AND ($2 = '' OR request_by = $2)
		AND ($3 = '' OR assigned_to = $3)
		AND ($4 = '' OR status = $4)
	`
	if err := r.db.QueryRowContext(ctx, countQuery, rbwID, requestBy, assignedTo, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, rbw_id, node_id, request_by, assigned_to, approved_by, type, status,
		       request_date, schedule_date, uninstall_date, issue, resolution, notes,
		       created_at, updated_at
		FROM service_requests 
		WHERE ($1 = '' OR rbw_id = $1)
		AND ($2 = '' OR request_by = $2)
		AND ($3 = '' OR assigned_to = $3)
		AND ($4 = '' OR status = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.db.QueryContext(ctx, query, rbwID, requestBy, assignedTo, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		sr := &models.ServiceRequest{}
		if err := rows.Scan(
			&sr.ID, &sr.RBWID, &sr.NodeID, &sr.RequestBy, &sr.AssignedTo, &sr.ApprovedBy,
			&sr.Type, &sr.Status, &sr.RequestDate, &sr.ScheduleDate, &sr.UninstallDate,
			&sr.Issue, &sr.Resolution, &sr.Notes, &sr.CreatedAt, &sr.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		srs = append(srs, sr)
	}

	return srs, total, rows.Err()
}

// TransactionRepository interface
type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetByID(ctx context.Context, id string) (*models.Transaction, error)
	Update(ctx context.Context, tx *models.Transaction) error
	ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Transaction, int, error)
	ListCategories(ctx context.Context) ([]*models.TransactionCategory, error)
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (rbw_id, category_id, amount, type, description, transaction_date, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		tx.RBWID, tx.CategoryID, tx.Amount, tx.Type, tx.Description, tx.TransactionDate, tx.CreatedBy,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	query := `
		SELECT id, rbw_id, category_id, amount, type, description, transaction_date,
		       created_by, created_at, updated_at
		FROM transactions WHERE id = $1
	`
	tx := &models.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.RBWID, &tx.CategoryID, &tx.Amount, &tx.Type, &tx.Description,
		&tx.TransactionDate, &tx.CreatedBy, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTransactionNotFound
	}
	return tx, err
}

func (r *transactionRepository) Update(ctx context.Context, tx *models.Transaction) error {
	query := `
		UPDATE transactions 
		SET category_id = $1, amount = $2, description = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		tx.CategoryID, tx.Amount, tx.Description, tx.ID,
	).Scan(&tx.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrTransactionNotFound
	}
	return err
}

func (r *transactionRepository) ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Transaction, int, error) {
	var txs []*models.Transaction
	var total int

	countQuery := `SELECT COUNT(*) FROM transactions WHERE rbw_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, rbwID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, rbw_id, category_id, amount, type, description, transaction_date,
		       created_by, created_at, updated_at
		FROM transactions 
		WHERE rbw_id = $1
		ORDER BY transaction_date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, rbwID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		tx := &models.Transaction{}
		if err := rows.Scan(
			&tx.ID, &tx.RBWID, &tx.CategoryID, &tx.Amount, &tx.Type, &tx.Description,
			&tx.TransactionDate, &tx.CreatedBy, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}

	return txs, total, rows.Err()
}

func (r *transactionRepository) ListCategories(ctx context.Context) ([]*models.TransactionCategory, error) {
	query := `SELECT id, name, type, description, created_at FROM transaction_categories ORDER BY type, name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.TransactionCategory
	for rows.Next() {
		cat := &models.TransactionCategory{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Type, &cat.Description, &cat.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}
