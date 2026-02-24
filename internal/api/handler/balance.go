package handler

import (
	"net/http"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/api/middleware"
	"github.com/BilalGunden-Insider/go-backend/internal/api/response"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/repository"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
	"github.com/google/uuid"
)

type BalanceHandler struct {
	balanceSvc *service.BalanceService
	txnRepo    repository.TransactionRepository
}

func NewBalanceHandler(balanceSvc *service.BalanceService, txnRepo repository.TransactionRepository) *BalanceHandler {
	return &BalanceHandler{balanceSvc: balanceSvc, txnRepo: txnRepo}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	callerID, _ := middleware.UserIDFromContext(r.Context())
	callerRole := middleware.RoleFromContext(r.Context())
	if callerID != userID && callerRole != models.RoleAdmin {
		response.Forbidden(w)
		return
	}
	amount, err := h.balanceSvc.GetBalance(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"user_id": userID.String(),
		"balance": amount.String(),
	})
}

func (h *BalanceHandler) GetBalanceAt(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	callerID, _ := middleware.UserIDFromContext(r.Context())
	callerRole := middleware.RoleFromContext(r.Context())
	if callerID != userID && callerRole != models.RoleAdmin {
		response.Forbidden(w)
		return
	}
	atStr := r.URL.Query().Get("time")
	if atStr == "" {
		response.Error(w, http.StatusBadRequest, "time query parameter is required (RFC3339)")
		return
	}
	at, err := time.Parse(time.RFC3339, atStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid time format, use RFC3339 (e.g. 2024-01-01T00:00:00Z)")
		return
	}
	amount, err := h.txnRepo.CalculateBalanceAt(r.Context(), userID, at)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"user_id": userID.String(),
		"balance": amount.String(),
		"at":      at.Format(time.RFC3339),
	})
}
