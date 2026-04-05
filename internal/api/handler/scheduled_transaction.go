package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/api/middleware"
	"github.com/BilalGunden-Insider/go-backend/internal/api/response"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ScheduledTransactionHandler struct {
	repo repository.ScheduledTransactionRepository
}

func NewScheduledTransactionHandler(repo repository.ScheduledTransactionRepository) *ScheduledTransactionHandler {
	return &ScheduledTransactionHandler{repo: repo}
}

func (h *ScheduledTransactionHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FromUserID  string `json:"from_user_id"`
		ToUserID    string `json:"to_user_id"`
		Amount      string `json:"amount"`
		Type        string `json:"type"`
		ScheduledAt string `json:"scheduled_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(body.Amount)
	if err != nil || !amount.IsPositive() {
		response.Error(w, http.StatusBadRequest, "invalid amount")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, body.ScheduledAt)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid scheduled_at, use RFC3339 format")
		return
	}
	if scheduledAt.Before(time.Now()) {
		response.Error(w, http.StatusBadRequest, "scheduled_at must be in the future")
		return
	}

	switch body.Type {
	case models.TxTypeCredit, models.TxTypeDebit, models.TxTypeTransfer:
	default:
		response.Error(w, http.StatusBadRequest, "invalid type: must be credit, debit, or transfer")
		return
	}

	var fromID, toID uuid.UUID
	if body.FromUserID != "" {
		fromID, err = uuid.Parse(body.FromUserID)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid from_user_id")
			return
		}
	}
	if body.ToUserID != "" {
		toID, err = uuid.Parse(body.ToUserID)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid to_user_id")
			return
		}
	}

	st := &models.ScheduledTransaction{
		ID:          uuid.New(),
		FromUserID:  fromID,
		ToUserID:    toID,
		Amount:      amount,
		Type:        body.Type,
		Status:      models.SchedStatusPending,
		ScheduledAt: scheduledAt,
		CreatedAt:   time.Now(),
	}

	if err := h.repo.Create(r.Context(), st); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, st)
}

func (h *ScheduledTransactionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Cancel(r.Context(), id); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduledTransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	callerID, _ := middleware.UserIDFromContext(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}

	list, err := h.repo.ListByUser(r.Context(), callerID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *ScheduledTransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	st, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}
	response.JSON(w, http.StatusOK, st)
}
