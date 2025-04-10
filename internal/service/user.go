package service

import (
	"context"
	"key-haven-back/internal/domain/user"
	"key-haven-back/internal/repository"
	"key-haven-back/internal/service/dto"
	"key-haven-back/pkg/secret"
)

type UserService interface {
	CreateUser(ctx context.Context, request *dto.CreateUserRequest) (*user.User, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	UpdatePassword(ctx context.Context, userID, password string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, request *dto.CreateUserRequest) (*user.User, error) {
	// Create user model from request
	u := dto.NewUser(request)

	// Hash the password
	hashedPassword, err := secret.HashPassword(request.Password)
	if err != nil {
		return nil, err
	}
	u.Password = hashedPassword

	// Save user to repository
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	// Don't return the password
	u.Password = ""
	return u, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u.Password = ""
	return u, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

func (s *userService) UpdatePassword(ctx context.Context, userID, password string) error {
	hashedPassword, err := secret.HashPassword(password)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
}
