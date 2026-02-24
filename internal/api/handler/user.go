package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BilalGunden-Insider/go-backend/internal/api/middleware"
	"github.com/BilalGunden-Insider/go-backend/internal/api/response"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
	"github.com/google/uuid"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	users, err := h.userSvc.List(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	callerID, _ := middleware.UserIDFromContext(r.Context())
	callerRole := middleware.RoleFromContext(r.Context())
	if callerID != id && callerRole != models.RoleAdmin {
		response.Forbidden(w)
		return
	}
	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w)
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	callerID, _ := middleware.UserIDFromContext(r.Context())
	callerRole := middleware.RoleFromContext(r.Context())
	if callerID != id && callerRole != models.RoleAdmin {
		response.Forbidden(w)
		return
	}
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.userSvc.Update(r.Context(), id, body.Username, body.Email)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.userSvc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
