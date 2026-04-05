package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionLimit struct {
	ID                uuid.UUID       `json:"id"`
	UserID            *uuid.UUID      `json:"user_id,omitempty"`
	MaxPerTransaction decimal.Decimal `json:"max_per_transaction"`
	MaxDailyAmount    decimal.Decimal `json:"max_daily_amount"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
