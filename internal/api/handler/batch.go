package handler

import (
	"encoding/json"
	"net/http"

	"github.com/BilalGunden-Insider/go-backend/internal/api/response"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BatchHandler struct {
	txnSvc *service.TransactionService
}

func NewBatchHandler(txnSvc *service.TransactionService) *BatchHandler {
	return &BatchHandler{txnSvc: txnSvc}
}

type batchItem struct {
	Type       string `json:"type"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Amount     string `json:"amount"`
}

func (h *BatchHandler) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Transactions []batchItem `json:"transactions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(body.Transactions) == 0 {
		response.Error(w, http.StatusBadRequest, "no transactions provided")
		return
	}
	if len(body.Transactions) > 100 {
		response.Error(w, http.StatusBadRequest, "max 100 transactions per batch")
		return
	}

	accepted, rejected := 0, 0

	for _, item := range body.Transactions {
		tx, err := parseBatchItem(item)
		if err != nil {
			rejected++
			continue
		}

		if h.txnSvc.SubmitAsync(tx) {
			accepted++
		} else {
			rejected++
		}
	}

	response.JSON(w, http.StatusAccepted, map[string]int{
		"accepted": accepted,
		"rejected": rejected,
		"total":    len(body.Transactions),
	})
}

func (h *BatchHandler) WorkerStats(w http.ResponseWriter, r *http.Request) {
	stats := h.txnSvc.GetWorkerStats()
	response.JSON(w, http.StatusOK, stats)
}

func parseBatchItem(item batchItem) (*models.Transaction, error) {
	amount, err := decimal.NewFromString(item.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, err
	}

	var fromID, toID uuid.UUID
	if item.FromUserID != "" {
		fromID, err = uuid.Parse(item.FromUserID)
		if err != nil {
			return nil, err
		}
	}
	if item.ToUserID != "" {
		toID, err = uuid.Parse(item.ToUserID)
		if err != nil {
			return nil, err
		}
	}

	tx := &models.Transaction{
		ID:         uuid.New(),
		FromUserID: fromID,
		ToUserID:   toID,
		Amount:     amount,
		Type:       item.Type,
		Status:     models.TxStatusPending,
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}
	return tx, nil
}

