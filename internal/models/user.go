package models

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username is required")
	}
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return errors.New("invalid email address")
	}
	if u.Role != RoleUser && u.Role != RoleAdmin {
		return errors.New("role must be 'user' or 'admin'")
	}
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
