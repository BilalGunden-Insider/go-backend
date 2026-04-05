package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	SchedStatusPending   = "pending"
	SchedStatusExecuted  = "executed"
	SchedStatusFailed    = "failed"
	SchedStatusCancelled = "cancelled"
)

type ScheduledTransaction struct {
	ID           uuid.UUID       `json:"id"`
	FromUserID   uuid.UUID       `json:"from_user_id"`
	ToUserID     uuid.UUID       `json:"to_user_id"`
	Amount       decimal.Decimal `json:"amount"`
	Type         string          `json:"type"`
	Status       string          `json:"status"`
	ScheduledAt  time.Time       `json:"scheduled_at"`
	ExecutedAt   *time.Time      `json:"executed_at,omitempty"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}
