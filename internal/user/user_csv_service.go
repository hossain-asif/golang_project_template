package user

import (
	"fmt"
	common_csv "go_project_structure/common_pkg/csv"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/dto"
	"go_project_structure/utils/authentication"
)

func (us *UserServiceImpl) ExportUsersAsCSV() (string, error) {
	users, err := us.userRepository.GetAll()
	if err != nil {
		return "", err
	}

	var userCSV []dto.UserCSV
	for _, user := range users {
		userCSV = append(userCSV, dto.UserCSV{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	fileName, err := common_csv.ExportToCSV("users", userCSV)
	if err != nil {
		return "", err
	}

	return fileName, nil

}

func (us *UserServiceImpl) CreateUserViaTnxUsingBatchProcessing(batch [][]string) error {
	fmt.Println("Creating user in user service using batch processing.")

	var users []*models.User

	for _, user := range batch {

		password, passwordErr := authentication.HashPassword(user[2])
		if passwordErr != nil {
			return passwordErr
		}

		users = append(users, &models.User{
			Name:     user[0],
			Email:    user[1],
			Password: password,
		})
	}

	message, err := us.userRepository.InsertViaTnxUsingBatchProcessing(users)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return err
	}

	fmt.Println(message)

	return nil
}
