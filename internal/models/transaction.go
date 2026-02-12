package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	TxTypeTransfer = "transfer"
	TxTypeCredit   = "credit"
	TxTypeDebit    = "debit"

	TxStatusPending    = "pending"
	TxStatusCompleted  = "completed"
	TxStatusFailed     = "failed"
	TxStatusRolledBack = "rolled_back"
)

type Transaction struct {
	ID         uuid.UUID       `json:"id"`
	FromUserID uuid.UUID       `json:"from_user_id"`
	ToUserID   uuid.UUID       `json:"to_user_id"`
	Amount     decimal.Decimal `json:"amount"`
	Type       string          `json:"type"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (t *Transaction) Validate() error {
	if t.FromUserID == uuid.Nil && t.Type != TxTypeCredit {
		return errors.New("from_user_id is required for non-credit transactions")
	}
	if t.ToUserID == uuid.Nil && t.Type != TxTypeDebit {
		return errors.New("to_user_id is required for non-debit transactions")
	}
	if !t.Amount.IsPositive() {
		return errors.New("amount must be positive")
	}
	switch t.Type {
	case TxTypeTransfer, TxTypeCredit, TxTypeDebit:
	default:
		return fmt.Errorf("invalid transaction type: %s", t.Type)
	}
	return nil
}

func (t *Transaction) Transition(newStatus string) error {
	switch t.Status {
	case TxStatusPending:
		if newStatus == TxStatusCompleted || newStatus == TxStatusFailed {
			t.Status = newStatus
			return nil
		}
	case TxStatusCompleted, TxStatusFailed:
		if newStatus == TxStatusRolledBack {
			t.Status = newStatus
			return nil
		}
	}
	return fmt.Errorf("cannot transition from %s to %s", t.Status, newStatus)
}
