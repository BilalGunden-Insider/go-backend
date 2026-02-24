package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BilalGunden-Insider/go-backend/internal/api/middleware"
	"github.com/BilalGunden-Insider/go-backend/internal/api/response"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	txnSvc  *service.TransactionService
	txnRepo repository.TransactionRepository
}

func NewTransactionHandler(txnSvc *service.TransactionService, txnRepo repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{txnSvc: txnSvc, txnRepo: txnRepo}
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string `json:"user_id"`
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	amount, err := decimal.NewFromString(body.Amount)
	if err != nil || !amount.IsPositive() {
		response.Error(w, http.StatusBadRequest, "invalid amount")
		return
	}
	tx, err := h.txnSvc.Credit(r.Context(), userID, amount)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string `json:"user_id"`
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	amount, err := decimal.NewFromString(body.Amount)
	if err != nil || !amount.IsPositive() {
		response.Error(w, http.StatusBadRequest, "invalid amount")
		return
	}
	tx, err := h.txnSvc.Debit(r.Context(), userID, amount)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FromUserID string `json:"from_user_id"`
		ToUserID   string `json:"to_user_id"`
		Amount     string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fromID, err := uuid.Parse(body.FromUserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid from_user_id")
		return
	}
	toID, err := uuid.Parse(body.ToUserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid to_user_id")
		return
	}
	amount, err := decimal.NewFromString(body.Amount)
	if err != nil || !amount.IsPositive() {
		response.Error(w, http.StatusBadRequest, "invalid amount")
		return
	}
	tx, err := h.txnSvc.Transfer(r.Context(), fromID, toID, amount)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	callerID, _ := middleware.UserIDFromContext(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	txns, err := h.txnRepo.ListByUser(r.Context(), callerID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, txns)
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	tx, err := h.txnRepo.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}
	response.JSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) RollbackTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid transaction id")
		return
	}
	if err := h.txnSvc.Rollback(r.Context(), id); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
