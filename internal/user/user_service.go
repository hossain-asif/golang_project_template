package user

import (
	"fmt"
	env "go_project_structure/config/env"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/db/repositories"
	"go_project_structure/utils/authentication"

	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	CreateUser(user *models.User) (string, error)
	LoginUser(email string, password string) (string, error)
	GetUserById(id string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(id string, username *string, email *string) (string, error)
	DeleteUser(id string) (string, error)
	PermanentlyDeleteUser(id string) (string, error)

	ExportUsersAsCSV() (string, error)
	CreateUserViaTnx(users [][]string) (string, error)
	CreateUserViaTnxUsingBatchProcessing(users [][]string) error
}

type UserServiceImpl struct {
	userRepository repositories.UserRepository
}

func NewUserService(_userRepository repositories.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: _userRepository,
	}
}

func (us *UserServiceImpl) CreateUser(user *models.User) (string, error) {
	fmt.Println("Creating user in user service.")

	password, hashErr := authentication.HashPassword(user.Password)
	if hashErr != nil {
		fmt.Printf("Error hashing password: %v\n", hashErr)
		return "", hashErr
	}
	user.Password = password

	message, err := us.userRepository.Create(user)

	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return "", err
	}
	return message, nil
}

func (us *UserServiceImpl) LoginUser(email string, password string) (string, error) {
	fmt.Println("Logging in user in user service.")
	user, err := us.userRepository.GetByEmail(email)
	if err != nil {
		fmt.Printf("Error fetching user by email: %v\n", err)
		return "", err
	}

	IsPasswordValid := authentication.CheckPasswordHash(password, user.Password)
	if !IsPasswordValid {
		fmt.Println("Invalid password provided.")
		return "", fmt.Errorf("invalid credentials")
	}

	payload := jwt.MapClaims{
		"email": user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, tokenErr := token.SignedString([]byte(env.GetString("JWT_SECRET", "default_secret_key")))
	if tokenErr != nil {
		fmt.Printf("Error signing JWT token: %v\n", tokenErr)
		return "", tokenErr
	}
	fmt.Println("User logged in successfully.")
	return tokenString, nil
}

func (us *UserServiceImpl) GetUserById(id string) (*models.User, error) {
	fmt.Println("Getting user by id in user service.")
	user, err := us.userRepository.GetByID(id)
	if err != nil {
		fmt.Printf("Error fetching user by id: %v\n", err)
		return nil, err
	}
	return user, nil
}

func (us *UserServiceImpl) GetAllUsers() ([]*models.User, error) {
	fmt.Println("Getting all users in user service.")
	users, err := us.userRepository.GetAll()
	if err != nil {
		fmt.Printf("Error fetching all users: %v\n", err)
		return nil, err
	}
	return users, nil
}

func (us *UserServiceImpl) UpdateUser(id string, username *string, email *string) (string, error) {
	fmt.Println("Updating user in user service.")

	message, err := us.userRepository.Update(id, username, email)
	if err != nil {
		fmt.Printf("Error updating user: %v\n", err)
		return "", err
	}

	return message, nil
}

func (us *UserServiceImpl) DeleteUser(id string) (string, error) {
	fmt.Println("Deleting user in user service.")

	message, err := us.userRepository.SoftDelete(id)
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
		return "", err
	}

	return message, nil
}

func (us *UserServiceImpl) PermanentlyDeleteUser(id string) (string, error) {
	fmt.Println("Permanently deleting user in user service.")

	message, err := us.userRepository.HardDelete(id)
	if err != nil {
		fmt.Printf("Error permanently deleting user: %v\n", err)
		return "", err
	}

	return message, nil
}

func (us *UserServiceImpl) CreateUserViaTnx(users [][]string) (string, error) {
	fmt.Println("Creating user in user service.")

	var messages [][]string

	for _, user := range users {

		password, hashErr := authentication.HashPassword(user[2])
		if hashErr != nil {
			fmt.Printf("Error hashing password: %v\n", hashErr)
			return "", hashErr
		}

		newUser := &models.User{
			Name:     user[0],
			Email:    user[1],
			Password: password,
		}

		message, err := us.userRepository.InsertViaTnx(newUser)
		if err != nil {
			fmt.Printf("Error creating user: %v\n", err)
			return "", err
		}
		messages = append(messages, []string{message})
	}

	return fmt.Sprintf("CSV uploaded successfully. messages: %s\n", messages), nil
}

