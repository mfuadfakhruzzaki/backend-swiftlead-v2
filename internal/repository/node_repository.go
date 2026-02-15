package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/swiftlead/backend-swiftlet/internal/models"
)

var (
	ErrNodeNotFound = errors.New("node not found")
)

// NodeRepository interface
type NodeRepository interface {
	Create(ctx context.Context, node *models.Node) error
	GetByID(ctx context.Context, id string) (*models.Node, error)
	GetByESP32UID(ctx context.Context, esp32UID string) (*models.Node, error)
	Update(ctx context.Context, node *models.Node) error
	Delete(ctx context.Context, id string) error
	ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Node, int, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateLastSeen(ctx context.Context, id string) error
	GetGatewayByRBW(ctx context.Context, rbwID string) (*models.Node, error)
	UpdateAudioState(ctx context.Context, id string, lmbState, nestState *bool) error
	UpdatePumpState(ctx context.Context, id string, pumpState bool) error
}

type nodeRepository struct {
	db *sql.DB
}

func NewNodeRepository(db *sql.DB) NodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) Create(ctx context.Context, node *models.Node) error {
	query := `
		INSERT INTO nodes (rbw_id, node_type, node_code, esp32_uid, has_audio, has_pump)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status_node, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		node.RBWID, node.NodeType, node.NodeCode, node.ESP32UID,
		node.HasAudio, node.HasPump,
	).Scan(&node.ID, &node.StatusNode, &node.CreatedAt, &node.UpdatedAt)

	return err
}

func (r *nodeRepository) GetByID(ctx context.Context, id string) (*models.Node, error) {
	query := `
		SELECT id, rbw_id, node_type, node_code, esp32_uid, status_node, last_seen,
		       has_audio, state_audio_lmb, state_audio_nest, has_pump, state_pump,
		       installed_at, uninstalled_at, created_at, updated_at
		FROM nodes WHERE id = $1
	`
	node := &models.Node{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID, &node.RBWID, &node.NodeType, &node.NodeCode, &node.ESP32UID,
		&node.StatusNode, &node.LastSeen, &node.HasAudio, &node.StateAudioLMB,
		&node.StateAudioNest, &node.HasPump, &node.StatePump,
		&node.InstalledAt, &node.UninstalledAt, &node.CreatedAt, &node.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNodeNotFound
	}
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (r *nodeRepository) GetByESP32UID(ctx context.Context, esp32UID string) (*models.Node, error) {
	// Normalize MAC address (remove colons)
	normalized := strings.ReplaceAll(esp32UID, ":", "")

	query := `
		SELECT id, rbw_id, node_type, node_code, esp32_uid, status_node, last_seen,
		       has_audio, state_audio_lmb, state_audio_nest, has_pump, state_pump,
		       installed_at, uninstalled_at, created_at, updated_at
		FROM nodes 
		WHERE REPLACE(esp32_uid, ':', '') = $1 OR esp32_uid = $2
	`
	node := &models.Node{}
	err := r.db.QueryRowContext(ctx, query, normalized, esp32UID).Scan(
		&node.ID, &node.RBWID, &node.NodeType, &node.NodeCode, &node.ESP32UID,
		&node.StatusNode, &node.LastSeen, &node.HasAudio, &node.StateAudioLMB,
		&node.StateAudioNest, &node.HasPump, &node.StatePump,
		&node.InstalledAt, &node.UninstalledAt, &node.CreatedAt, &node.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNodeNotFound
	}
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (r *nodeRepository) Update(ctx context.Context, node *models.Node) error {
	query := `
		UPDATE nodes 
		SET node_code = $1, esp32_uid = $2, has_audio = $3, has_pump = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		node.NodeCode, node.ESP32UID, node.HasAudio, node.HasPump, node.ID,
	).Scan(&node.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNodeNotFound
	}
	return err
}

func (r *nodeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM nodes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNodeNotFound
	}
	return nil
}

func (r *nodeRepository) ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Node, int, error) {
	var nodes []*models.Node
	var total int

	countQuery := `SELECT COUNT(*) FROM nodes WHERE rbw_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, rbwID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, rbw_id, node_type, node_code, esp32_uid, status_node, last_seen,
		       has_audio, state_audio_lmb, state_audio_nest, has_pump, state_pump,
		       installed_at, uninstalled_at, created_at, updated_at
		FROM nodes 
		WHERE rbw_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, rbwID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		node := &models.Node{}
		if err := rows.Scan(
			&node.ID, &node.RBWID, &node.NodeType, &node.NodeCode, &node.ESP32UID,
			&node.StatusNode, &node.LastSeen, &node.HasAudio, &node.StateAudioLMB,
			&node.StateAudioNest, &node.HasPump, &node.StatePump,
			&node.InstalledAt, &node.UninstalledAt, &node.CreatedAt, &node.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		nodes = append(nodes, node)
	}

	return nodes, total, rows.Err()
}

func (r *nodeRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE nodes SET status_node = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *nodeRepository) UpdateLastSeen(ctx context.Context, id string) error {
	query := `UPDATE nodes SET last_seen = NOW(), status_node = 'online', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *nodeRepository) GetGatewayByRBW(ctx context.Context, rbwID string) (*models.Node, error) {
	query := `
		SELECT id, rbw_id, node_type, node_code, esp32_uid, status_node, last_seen,
		       has_audio, state_audio_lmb, state_audio_nest, has_pump, state_pump,
		       installed_at, uninstalled_at, created_at, updated_at
		FROM nodes
		WHERE rbw_id = $1 AND node_type = 'gateway'
		LIMIT 1
	`
	node := &models.Node{}
	err := r.db.QueryRowContext(ctx, query, rbwID).Scan(
		&node.ID, &node.RBWID, &node.NodeType, &node.NodeCode, &node.ESP32UID,
		&node.StatusNode, &node.LastSeen, &node.HasAudio, &node.StateAudioLMB,
		&node.StateAudioNest, &node.HasPump, &node.StatePump,
		&node.InstalledAt, &node.UninstalledAt, &node.CreatedAt, &node.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No gateway registered for this RBW
	}
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (r *nodeRepository) UpdateAudioState(ctx context.Context, id string, lmbState, nestState *bool) error {
	query := `
		UPDATE nodes 
		SET state_audio_lmb = COALESCE($1, state_audio_lmb),
		    state_audio_nest = COALESCE($2, state_audio_nest),
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, lmbState, nestState, id)
	return err
}

func (r *nodeRepository) UpdatePumpState(ctx context.Context, id string, pumpState bool) error {
	query := `UPDATE nodes SET state_pump = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, pumpState, id)
	return err
}
