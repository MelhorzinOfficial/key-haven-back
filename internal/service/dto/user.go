package dto

import (
	"key-haven-back/internal/domain/user"
	"time"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
	Name     string `json:"name" validate:"required"`
}

type CreateUserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name" `
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string    `json:"token"`
	User  user.User `json:"user"`
}

func NewUser(req *CreateUserRequest) *user.User {
	now := time.Now()
	return &user.User{
		ID:          uuid.New().String(),
		Email:       req.Email,
		Name:        req.Name,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: now,
	}
}
