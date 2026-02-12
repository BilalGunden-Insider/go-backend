package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bilal/backend_path/internal/models"
	"github.com/bilal/backend_path/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users    repository.UserRepository
	balances repository.BalanceRepository
	audit    repository.AuditLogRepository
	log      *slog.Logger
}

func NewUserService(
	users repository.UserRepository,
	balances repository.BalanceRepository,
	audit repository.AuditLogRepository,
	log *slog.Logger,
) *UserService {
	return &UserService{users: users, balances: balances, audit: audit, log: log}
}

func (s *UserService) Register(ctx context.Context, username, email, password, role string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	balance := &models.Balance{
		UserID:        user.ID,
		Amount:        decimal.Zero,
		LastUpdatedAt: now,
	}
	if err := s.balances.Create(ctx, balance); err != nil {
		return nil, fmt.Errorf("create balance: %w", err)
	}

	details, _ := json.Marshal(map[string]string{"username": username, "email": email, "role": role})
	_ = s.audit.Create(ctx, &models.AuditLog{
		ID:         uuid.New(),
		EntityType: models.EntityUser,
		EntityID:   user.ID,
		Action:     models.ActionCreate,
		Details:    details,
		CreatedAt:  now,
	})

	s.log.Info("user registered", slog.String("user_id", user.ID.String()), slog.String("username", username))
	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (s *UserService) Authorize(user *models.User, requiredRole string) error {
	if requiredRole == models.RoleAdmin && !user.IsAdmin() {
		return errors.New("admin access required")
	}
	return nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.users.GetByID(ctx, id)
}
