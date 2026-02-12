package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	EntityUser        = "user"
	EntityTransaction = "transaction"
	EntityBalance     = "balance"

	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionTransfer = "transfer"
	ActionRollback = "rollback"
)

type AuditLog struct {
	ID         uuid.UUID       `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id"`
	Action     string          `json:"action"`
	Details    json.RawMessage `json:"details,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
