package services

import (
	"context"
	"testing"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

// --- Mock NodeRepository for pump tests ---

type MockNodeRepository struct {
	nodes map[string]*models.Node
	// Track calls for assertions
	PumpStateUpdates    []pumpStateUpdate
	PumpAutoModeUpdates []pumpAutoModeUpdate
}

type pumpStateUpdate struct {
	NodeID string
	State  bool
}

type pumpAutoModeUpdate struct {
	NodeID   string
	AutoMode bool
}

func NewMockNodeRepository() *MockNodeRepository {
	return &MockNodeRepository{
		nodes: make(map[string]*models.Node),
	}
}

func (m *MockNodeRepository) AddNode(node *models.Node) {
	m.nodes[node.ID] = node
}

func (m *MockNodeRepository) GetByID(ctx context.Context, id string) (*models.Node, error) {
	if node, ok := m.nodes[id]; ok {
		return node, nil
	}
	return nil, repository.ErrNodeNotFound
}

func (m *MockNodeRepository) UpdatePumpState(ctx context.Context, id string, pumpState bool) error {
	m.PumpStateUpdates = append(m.PumpStateUpdates, pumpStateUpdate{id, pumpState})
	if node, ok := m.nodes[id]; ok {
		node.StatePump = &pumpState
	}
	return nil
}

func (m *MockNodeRepository) UpdatePumpAutoMode(ctx context.Context, id string, autoMode bool) error {
	m.PumpAutoModeUpdates = append(m.PumpAutoModeUpdates, pumpAutoModeUpdate{id, autoMode})
	if node, ok := m.nodes[id]; ok {
		node.PumpAutoMode = autoMode
	}
	return nil
}

// Unused interface methods — stubs
func (m *MockNodeRepository) Create(ctx context.Context, node *models.Node) error { return nil }
func (m *MockNodeRepository) GetByESP32UID(ctx context.Context, esp32UID string) (*models.Node, error) {
	return nil, repository.ErrNodeNotFound
}
func (m *MockNodeRepository) Update(ctx context.Context, node *models.Node) error { return nil }
func (m *MockNodeRepository) Delete(ctx context.Context, id string) error         { return nil }
func (m *MockNodeRepository) ListByRBW(ctx context.Context, rbwID string, limit, offset int) ([]*models.Node, int, error) {
	return nil, 0, nil
}
func (m *MockNodeRepository) UpdateStatus(ctx context.Context, id, status string) error { return nil }
func (m *MockNodeRepository) UpdateLastSeen(ctx context.Context, id string) error       { return nil }
func (m *MockNodeRepository) GetGatewayByRBW(ctx context.Context, rbwID string) (*models.Node, error) {
	return nil, nil
}
func (m *MockNodeRepository) UpdateAudioState(ctx context.Context, id string, lmbState, nestState *bool) error {
	return nil
}
func (m *MockNodeRepository) UpdateLastSeenByRBWAndTypes(ctx context.Context, rbwID string, nodeTypes []string) error {
	return nil
}
func (m *MockNodeRepository) ListStale(ctx context.Context, olderThan time.Time) ([]*models.Node, error) {
	return nil, nil
}
func (m *MockNodeRepository) GetPumpNodesByRBW(ctx context.Context, rbwID string) ([]*models.Node, error) {
	var nodes []*models.Node
	for _, n := range m.nodes {
		if n.RBWID == rbwID && n.HasPump {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}
func (m *MockNodeRepository) ListAllWithPump(ctx context.Context) ([]*models.Node, error) {
	var nodes []*models.Node
	for _, n := range m.nodes {
		if n.HasPump {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

// --- Tests for GetPumpNode ---

func TestGetPumpNode_Success(t *testing.T) {
	repo := NewMockNodeRepository()
	svc := NewAudioService(nil, repo)

	repo.AddNode(&models.Node{
		ID:      "pump-1",
		HasPump: true,
	})

	node, err := svc.GetPumpNode(context.Background(), "pump-1")
	if err != nil {
		t.Fatalf("GetPumpNode failed: %v", err)
	}
	if node.ID != "pump-1" {
		t.Errorf("Expected node ID pump-1, got %s", node.ID)
	}
}

func TestGetPumpNode_NoPumpCapability(t *testing.T) {
	repo := NewMockNodeRepository()
	svc := NewAudioService(nil, repo)

	repo.AddNode(&models.Node{
		ID:      "nest-1",
		HasPump: false,
	})

	_, err := svc.GetPumpNode(context.Background(), "nest-1")
	if err == nil {
		t.Fatal("Expected error for node without pump capability")
	}
}

func TestGetPumpNode_NotFound(t *testing.T) {
	repo := NewMockNodeRepository()
	svc := NewAudioService(nil, repo)

	_, err := svc.GetPumpNode(context.Background(), "nonexistent")
	if err != repository.ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

// --- Tests for ControlPumpAI mode isolation ---

func TestControlPumpAI_ManualMode_NoOp(t *testing.T) {
	repo := NewMockNodeRepository()
	svc := NewAudioService(nil, repo)

	pumpOff := false
	repo.AddNode(&models.Node{
		ID:           "pump-1",
		HasPump:      true,
		PumpAutoMode: false, // manual mode
		StatePump:    &pumpOff,
	})

	req := &models.PumpControlRequest{Action: "sprayer_set", Value: 1}
	err := svc.ControlPumpAI(context.Background(), "pump-1", req)
	if err != nil {
		t.Fatalf("ControlPumpAI should return nil in manual mode, got: %v", err)
	}

	// Verify no pump state updates happened (no-op)
	if len(repo.PumpStateUpdates) != 0 {
		t.Errorf("Expected no pump state updates in manual mode, got %d", len(repo.PumpStateUpdates))
	}
}

func TestControlPumpAI_AIMode_Actuates(t *testing.T) {
	repo := NewMockNodeRepository()

	pumpOff := false
	repo.AddNode(&models.Node{
		ID:           "pump-1",
		HasPump:      true,
		PumpAutoMode: true, // AI mode
		StatePump:    &pumpOff,
	})

	// We can't easily mock *mqtt.Client, so we verify the logic path:
	// When PumpAutoMode=true, ControlPumpAI should NOT return nil as a no-op.
	// It should attempt to call ControlPump (which will panic with nil mqtt).
	// We use recover to catch the panic — the panic itself proves the AI path was taken.
	svc := NewAudioService(nil, repo)

	req := &models.PumpControlRequest{Action: "sprayer_set", Value: 1}

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		svc.ControlPumpAI(context.Background(), "pump-1", req)
	}()

	// The panic proves ControlPumpAI tried to call ControlPump → PublishPumpCommand
	// (which panics on nil mqtt client). If it were in manual mode, it would
	// return nil immediately without reaching the MQTT publish.
	if !didPanic {
		t.Error("Expected ControlPumpAI to attempt actuation (reach MQTT publish) in AI mode")
	}
}

// --- Tests for SetPumpAutoMode ---

func TestSetPumpAutoMode_ToManual(t *testing.T) {
	repo := NewMockNodeRepository()
	svc := NewAudioService(nil, repo)

	repo.AddNode(&models.Node{
		ID:           "pump-1",
		HasPump:      true,
		PumpAutoMode: true,
	})

	err := svc.SetPumpAutoMode(context.Background(), "pump-1", false)
	if err != nil {
		t.Fatalf("SetPumpAutoMode failed: %v", err)
	}

	if len(repo.PumpAutoModeUpdates) != 1 {
		t.Fatalf("Expected 1 auto mode update, got %d", len(repo.PumpAutoModeUpdates))
	}
	if repo.PumpAutoModeUpdates[0].AutoMode != false {
		t.Error("Expected auto_mode to be set to false")
	}
}

func TestSetPumpAutoMode_ToAI(t *testing.T) {
	repo := NewMockNodeRepository()
	svc := NewAudioService(nil, repo)

	repo.AddNode(&models.Node{
		ID:           "pump-1",
		HasPump:      true,
		PumpAutoMode: false,
	})

	err := svc.SetPumpAutoMode(context.Background(), "pump-1", true)
	if err != nil {
		t.Fatalf("SetPumpAutoMode failed: %v", err)
	}

	if len(repo.PumpAutoModeUpdates) != 1 {
		t.Fatalf("Expected 1 auto mode update, got %d", len(repo.PumpAutoModeUpdates))
	}
	if repo.PumpAutoModeUpdates[0].AutoMode != true {
		t.Error("Expected auto_mode to be set to true")
	}
}
