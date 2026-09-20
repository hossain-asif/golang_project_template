package identity

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/db/repositories/user"
	"go_project_structure/internal/dto/identity"
	"go_project_structure/utils/authentication"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	CreateUser(ctx context.Context, user *models.User) (string, error)
	LoginUser(ctx context.Context, loginPayload *identityDTO.LoginUserRequest) (string, error)
	GetUserById(ctx context.Context, id string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, id string, updatePayload *identityDTO.UpdateUserRequest) (string, error)
	DeleteUser(ctx context.Context, id string) (string, error)
	PermanentlyDeleteUser(ctx context.Context, id string) (string, error)
}

type ServiceImpl struct {
	userRepository user.Repository
	Log            *logger.ScopeLogger
}

func NewService(_userRepository user.Repository) Service {
	return &ServiceImpl{
		userRepository: _userRepository,
		Log:            logger.Log.Scope("", "user", "user_service"),
	}
}

func (us *ServiceImpl) CreateUser(ctx context.Context, user *models.User) (string, error) {
	log := us.Log.Method("CreateUser").WithContext(ctx)

	password, hashErr := authentication.HashPassword(user.Password)
	if hashErr != nil {
		log.Errorf("Error hashing password: %v\n", hashErr)
		return "", hashErr
	}
	user.Password = password

	message, err := us.userRepository.Create(ctx, user)
	if err != nil {
		log.Errorf("Error creating user: %v\n", err)
		return "", err
	}

	return message, nil
}

func (us *ServiceImpl) LoginUser(ctx context.Context, loginPayload *identityDTO.LoginUserRequest) (string, error) {
	log := us.Log.Method("LoginUser").WithContext(ctx)

	user, err := us.userRepository.GetByEmail(ctx, loginPayload.Email)
	if err != nil {
		log.Errorf("Error fetching user by email: %v\n", err)
		return "", err
	}

	IsPasswordValid := authentication.CheckPasswordHash(loginPayload.Password, user.Password)
	if !IsPasswordValid {
		log.Errorf("Invalid password provided.")
		return "", fmt.Errorf("invalid credentials")
	}

	payload := jwt.MapClaims{
		"email": user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, tokenErr := token.SignedString([]byte(env.GetString("JWT_SECRET", "default_secret_key")))
	if tokenErr != nil {
		log.Errorf("Error signing JWT token: %v\n", tokenErr)
		return "", tokenErr
	}

	return tokenString, nil
}

func (us *ServiceImpl) GetUserById(ctx context.Context, id string) (*models.User, error) {
	log := us.Log.Method("GetUserById").WithContext(ctx)

	user, err := us.userRepository.GetByID(ctx, id)
	if err != nil {
		log.Errorf("Error fetching user by id: %v\n", err)
		return nil, err
	}

	return user, nil
}

func (us *ServiceImpl) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	log := us.Log.Method("GetAllUsers").WithContext(ctx)

	users, err := us.userRepository.GetAll(ctx)
	if err != nil {
		log.Errorf("Error fetching all users: %v\n", err)
		return nil, err
	}

	return users, nil
}

func (us *ServiceImpl) UpdateUser(ctx context.Context, id string, updatePayload *identityDTO.UpdateUserRequest) (string, error) {
	log := us.Log.Method("UpdateUser").WithContext(ctx)

	message, err := us.userRepository.Update(ctx, id, updatePayload)
	if err != nil {
		log.Errorf("Error updating user: %v\n", err)
		return "", err
	}

	return message, nil
}

func (us *ServiceImpl) DeleteUser(ctx context.Context, id string) (string, error) {
	log := us.Log.Method("DeleteUser").WithContext(ctx)

	message, err := us.userRepository.SoftDelete(ctx, id)
	if err != nil {
		log.Errorf("Error deleting user: %v\n", err)
		return "", err
	}

	return message, nil
}

func (us *ServiceImpl) PermanentlyDeleteUser(ctx context.Context, id string) (string, error) {
	log := us.Log.Method("PermanentlyDeleteUser").WithContext(ctx)

	message, err := us.userRepository.HardDelete(ctx, id)
	if err != nil {
		log.Errorf("Error permanently deleting user: %v\n", err)
		return "", err
	}

	return message, nil
}
