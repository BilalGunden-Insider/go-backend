package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Balance struct {
	UserID        uuid.UUID       `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	LastUpdatedAt time.Time       `json:"last_updated_at"`
}

func (b *Balance) Validate() error {
	if b.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	return nil
}

func (b *Balance) HasSufficientFunds(amount decimal.Decimal) bool {
	return b.Amount.GreaterThanOrEqual(amount)
}
