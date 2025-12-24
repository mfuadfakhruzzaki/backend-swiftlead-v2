package models

import "time"

// ServiceRequest represents a service request
type ServiceRequest struct {
	ID            string     `json:"id"`
	RBWID         string     `json:"rbw_id"`
	NodeID        *string    `json:"node_id,omitempty"`
	RequestBy     string     `json:"request_by"`
	AssignedTo    *string    `json:"assigned_to,omitempty"`
	ApprovedBy    *string    `json:"approved_by,omitempty"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	RequestDate   time.Time  `json:"request_date"`
	ScheduleDate  *time.Time `json:"schedule_date,omitempty"`
	UninstallDate *time.Time `json:"uninstall_date,omitempty"`
	Issue         *string    `json:"issue,omitempty"`
	Resolution    *string    `json:"resolution,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ServiceType constants
const (
	ServiceTypeInstallation = "installation"
	ServiceTypeMaintenance  = "maintenance"
	ServiceTypeUninstall    = "uninstall"
)

// ServiceStatus constants
const (
	ServiceStatusDraft      = "draft"
	ServiceStatusPending    = "pending"
	ServiceStatusApproved   = "approved"
	ServiceStatusRejected   = "rejected"
	ServiceStatusAssigned   = "assigned"
	ServiceStatusInProgress = "in_progress"
	ServiceStatusResolved   = "resolved"
	ServiceStatusCancelled  = "cancelled"
)

// CreateServiceRequestRequest for creating a service request
type CreateServiceRequestRequest struct {
	RBWID  string  `json:"rbw_id"`
	NodeID *string `json:"node_id,omitempty"`
	Type   string  `json:"type"`
	Issue  string  `json:"issue,omitempty"`
}

// UpdateServiceRequestRequest for updating a service request
type UpdateServiceRequestRequest struct {
	Status       *string    `json:"status,omitempty"`
	AssignedTo   *string    `json:"assigned_to,omitempty"`
	ScheduleDate *time.Time `json:"schedule_date,omitempty"`
	Resolution   *string    `json:"resolution,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}
