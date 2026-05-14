package user

import (
	"context"
	common_csv "go_project_structure/common_pkg/csv"
	"go_project_structure/internal/dto"
	"go_project_structure/internal/db/models"
	"go_project_structure/utils/authentication"
)

func (us *UserServiceImpl) ExportUsersAsCSV(ctx context.Context) (string, error) {
	log := us.userServiceLog.Method("ExportUsersAsCSV").WithContext(ctx)
	log.Infof("Exporting users as CSV in user service.")

	users, err := us.userRepository.GetAll(ctx)
	if err != nil {
		log.Errorf("Error fetching all users: %v\n", err)
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
		log.Errorf("Error exporting users as CSV: %v\n", err)
		return "", err
	}

	log.Infof("Users exported as CSV successfully from service. filename: %v", fileName)
	return fileName, nil

}

func (us *UserServiceImpl) CreateUserViaTnxUsingBatchProcessing(ctx context.Context, batch [][]string) error {
	log := us.userServiceLog.Method("CreateUserViaTnxUsingBatchProcessing").WithContext(ctx)
	log.Infof("Creating user in user service using batch processing.")

	var users []*models.User

	for _, user := range batch {

		password, passwordErr := authentication.HashPassword(user[2])
		if passwordErr != nil {
			log.Errorf("Error hashing password: %v\n", passwordErr)
			return passwordErr
		}

		users = append(users, &models.User{
			Name:     user[0],
			Email:    user[1],
			Password: password,
		})
	}

	message, err := us.userRepository.InsertViaTnxUsingBatchProcessing(ctx, users)
	if err != nil {
		log.Errorf("Error creating user: %v\n", err)
		return err
	}

	log.WithFields(map[string]interface{}{
		"message": message,
	}).Infof("User created via tnx successfully from service using batch processing.")

	return nil
}
