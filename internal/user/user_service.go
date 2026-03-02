package user

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"
	"go_project_structure/internal/infrastructure/models"
	"go_project_structure/internal/infrastructure/repositories"
	"go_project_structure/internal/dto"
	"go_project_structure/utils/authentication"

	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (string, error)
	LoginUser(ctx context.Context, loginPayload *dto.LoginUserRequest) (string, error)
	GetUserById(ctx context.Context, id string) (*models.User, error)
	GetAllUsers(ctx context.Context,) ([]*models.User, error)
	UpdateUser(ctx context.Context, id string, updatePayload *dto.UpdateUserRequest) (string, error)
	DeleteUser(ctx context.Context, id string) (string, error)
	PermanentlyDeleteUser(ctx context.Context, id string) (string, error)

	ExportUsersAsCSV(ctx context.Context) (string, error)
	CreateUserViaTnx(ctx context.Context, users [][]string) (string, error)
	CreateUserViaTnxUsingBatchProcessing(ctx context.Context, users [][]string) error
}

type UserServiceImpl struct {
	userRepository repositories.UserRepository
	userServiceLog *logger.ScopeLogger
}

func NewUserService(_userRepository repositories.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: _userRepository,
		userServiceLog: logger.Log.Scope("", "user", "user_service"),
	}
}

func (us *UserServiceImpl) CreateUser(ctx context.Context, user *models.User) (string, error) {
	log := us.userServiceLog.Method("CreateUser").WithContext(ctx)
	log.Info("Creating user in user service.")

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

	log.Info("User created successfully from service.")
	return message, nil
}

func (us *UserServiceImpl) LoginUser(ctx context.Context, loginPayload *dto.LoginUserRequest) (string, error) {
	log := us.userServiceLog.Method("LoginUser").WithContext(ctx)
	log.Info("Logging in user in user service.")

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
	log.Info("Login successful from service.")
	return tokenString, nil
}

func (us *UserServiceImpl) GetUserById(ctx context.Context, id string) (*models.User, error) {
	log := us.userServiceLog.Method("GetUserById").WithContext(ctx)
	log.Infof("Get user by id start.")

	user, err := us.userRepository.GetByID(ctx, id)
	if err != nil {
		log.Errorf("Error fetching user by id: %v\n", err)
		return nil, err
	}

	log.Infof("Get user by id from service.")
	return user, nil
}

func (us *UserServiceImpl) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	log := us.userServiceLog.Method("GetAllUsers").WithContext(ctx)
	log.Infof("Get all users start.")

	users, err := us.userRepository.GetAll(ctx)
	if err != nil {
		log.Errorf("Error fetching all users: %v\n", err)
		return nil, err
	}

	log.Infof("Get all users from service.")
	return users, nil
}

func (us *UserServiceImpl) UpdateUser(ctx context.Context, id string, updatePayload *dto.UpdateUserRequest) (string, error) {
	log := us.userServiceLog.Method("UpdateUser").WithContext(ctx)
	log.Infof("Update user start.")

	message, err := us.userRepository.Update(ctx, id, updatePayload)
	if err != nil {
		log.Errorf("Error updating user: %v\n", err)
		return "", err
	}

	log.Infof("Update user from service.")
	return message, nil
}

func (us *UserServiceImpl) DeleteUser(ctx context.Context, id string) (string, error) {
	log := us.userServiceLog.Method("DeleteUser").WithContext(ctx)
	log.Infof("Delete user start.")

	message, err := us.userRepository.SoftDelete(ctx, id)
	if err != nil {
		log.Errorf("Error deleting user: %v\n", err)
		return "", err
	}

	log.Infof("Delete user from service.")
	return message, nil
}

func (us *UserServiceImpl) PermanentlyDeleteUser(ctx context.Context, id string) (string, error) {
	log := us.userServiceLog.Method("PermanentlyDeleteUser").WithContext(ctx)
	log.Infof("Permanently delete user start.")

	message, err := us.userRepository.HardDelete(ctx, id)
	if err != nil {
		log.Errorf("Error permanently deleting user: %v\n", err)
		return "", err
	}

	log.Infof("Permanently delete user from service.")
	return message, nil
}

func (us *UserServiceImpl) CreateUserViaTnx(ctx context.Context, users [][]string) (string, error) {
	log := us.userServiceLog.Method("CreateUserViaTnx").WithContext(ctx)
	log.Info("Creating user in user service.")

	var messages [][]string

	for _, user := range users {

		password, hashErr := authentication.HashPassword(user[2])
		if hashErr != nil {
			log.Errorf("Error hashing password: %v\n", hashErr)
			return "", hashErr
		}

		newUser := &models.User{
			Name:     user[0],
			Email:    user[1],
			Password: password,
		}

		message, err := us.userRepository.InsertViaTnx(ctx, newUser)
		if err != nil {
			log.Errorf("Error creating user: %v\n", err)
			return "", err
		}
		messages = append(messages, []string{message})
	}

	log.Info("User created via tnx successfully from service.")
	return fmt.Sprintf("CSV uploaded successfully. messages: %s\n", messages), nil
}

