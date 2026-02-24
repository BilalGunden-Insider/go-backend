package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/BilalGunden-Insider/go-backend/internal/api/response"
	"github.com/BilalGunden-Insider/go-backend/internal/auth"
	"github.com/BilalGunden-Insider/go-backend/internal/models"
	"github.com/BilalGunden-Insider/go-backend/internal/service"
)

type AuthHandler struct {
	userSvc   *service.UserService
	jwtSecret string
}

func NewAuthHandler(userSvc *service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Role == "" {
		body.Role = models.RoleUser
	}
	user, err := h.userSvc.Register(r.Context(), body.Username, body.Email, body.Password, body.Role)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.userSvc.Authenticate(r.Context(), body.Email, body.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	token, err := auth.GenerateToken(user.ID, user.Role, h.jwtSecret, 24*time.Hour)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"token": token})
}
