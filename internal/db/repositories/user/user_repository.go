package repositories

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/db/models"
	"go_project_structure/internal/dto"
	"go_project_structure/utils/pg"
	"strings"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (string, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, id string, updatePayload *dto.UpdateUserRequest) (string, error)
	SoftDelete(ctx context.Context, id string) (string, error)
	HardDelete(ctx context.Context, id string) (string, error)

	InsertViaTnx(ctx context.Context, user *models.User) (string, error)
	InsertViaTnxUsingBatchProcessing(ctx context.Context, users []*models.User) (string, error)

	// user specific methods
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// pagination specific methods
	PaginationRepository
}

type UserRepositoryImpl struct {
	// Add fields for database connection, etc.
	db                *gorm.DB
	userRepositoryLog *logger.ScopeLogger
}

func NewUserRepository(_db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{
		db:                _db,
		userRepositoryLog: logger.Log.Scope("db", "repositories", "user_repository"),
	}
}

func (u *UserRepositoryImpl) Create(ctx context.Context, user *models.User) (string, error) {
	log := u.userRepositoryLog.Method("Create").WithContext(ctx)

	// step 1: prepare the query
	query := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"

	// step 2: execute the query
	result := u.db.Exec(query, user.Name, user.Email, user.Password)
	u.db.Debug()

	// step 3: check for errors
	if result.Error != nil {
		err := pg.HandlePgError(result.Error)

		log.Errorf("Error creating user: %v\n", err)

		return "", err
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {

		log.Error("No user was created.")

		return "", fmt.Errorf("No user was created.")
	}

	// step 5: return the result
	return fmt.Sprintf("Created user (rows affected: %d)\n", rowsAffected), nil
}

func (u *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*models.User, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("GetByID")

	// step 1: prepare the query
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE deleted_at IS NULL AND id = ?"

	// step 2: execute the query
	row := u.db.Raw(query, id).Row()

	// step 3: process the result
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorf("User not found.")
			return nil, err
		}
		log.Errorf("Error fetching user: %v\n", err)
		return nil, err
	}

	// step 4: return the result
	return user, nil
}

func (u *UserRepositoryImpl) GetAll(ctx context.Context) ([]*models.User, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("GetAll")

	// step 1: prepare the query
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE deleted_at IS NULL"

	// step 2: execute the query
	rows, err := u.db.Raw(query).Rows()
	if err != nil {
		log.Errorf("Error executing query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	// another way

	// step 2: execute the query with Raw() - NOT Exec()
	// var users []*models.User
	// result := u.db.Raw(query).Scan(&users)

	// step 3: check for errors
	// if result.Error != nil {
	//     fmt.Printf("Error fetching users: %v\n", result.Error)
	//     return nil, result.Error
	// }

	// step 4: process the result
	var users []*models.User
	for rows.Next() {
		var user models.User
		err := u.db.ScanRows(rows, &user)
		if err != nil {
			log.Errorf("Error scanning row: %v\n", err)
			return nil, err
		}
		users = append(users, &user)
	}

	// step 5: return the result
	return users, nil
}

func (u *UserRepositoryImpl) Update(ctx context.Context, id string, updatePayload *dto.UpdateUserRequest) (string, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("Update")

	// step 1: prepare the query
	query := "UPDATE users SET "
	args := []interface{}{}
	if updatePayload.Name != nil {
		query += "name = ?, "
		args = append(args, *updatePayload.Name)
	}
	if updatePayload.Email != nil {
		query += "email = ?, "
		args = append(args, *updatePayload.Email)
	}
	query += "updated_at = NOW() "
	query += "WHERE deleted_at IS NULL AND id = ?"
	args = append(args, id)

	// step 2: execute the query
	result := u.db.Exec(query, args...)

	// step 3: check for errors
	if result.Error != nil {
		log.Errorf("Error updating user: %v\n", result.Error)
		return "", result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		log.Errorf("No user was updated.")
		return "", fmt.Errorf("No user was updated.")
	}

	// step 5: return the result
	return fmt.Sprintf("User updated successfully (rows affected: %d)", rowsAffected), nil
}

func (u *UserRepositoryImpl) SoftDelete(ctx context.Context, id string) (string, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("SoftDelete")

	// step 1: prepare the query
	query := "UPDATE users SET deleted_at = NOW() WHERE deleted_at IS NULL AND id = ?"

	// step 2: execute the query
	result := u.db.Exec(query, id)

	// step 3: check for errors
	if result.Error != nil {
		log.Errorf("Error deleting user: %v\n", result.Error)
		return "", result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		log.Errorf("No user was deleted.")
		return "", fmt.Errorf("No user was deleted.")
	}

	// step 5: return the result
	return fmt.Sprintf("Deleted user (rows affected: %d)\n", rowsAffected), nil
}

func (u *UserRepositoryImpl) HardDelete(ctx context.Context, id string) (string, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("HardDelete")

	// step 1: prepare the query
	query := "DELETE FROM users WHERE id = ?"

	// step 2: execute the query
	result := u.db.Exec(query, id)

	// step 3: check for errors
	if result.Error != nil {
		log.Errorf("Error deleting user: %v\n", result.Error)
		return "", result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		log.Errorf("No user was deleted.")
		return "", fmt.Errorf("No user was deleted.")
	}

	// step 5: return the result
	return fmt.Sprintf("Deleted user (rows affected: %d)\n", rowsAffected), nil
}

func (u *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	log := u.userRepositoryLog.Method("GetByEmail").WithContext(ctx)

	// step 1: prepare the query
	query := "SELECT name, email, password FROM users WHERE deleted_at IS NULL AND email = ?"

	// step 2: execute the query
	row := u.db.Raw(query, email).Row()

	// step 3: process the result
	user := &models.User{}
	err := row.Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorf("User not found.")
			return nil, err
		}
		log.Errorf("Error fetching user: %v\n", err)
		return nil, err
	}

	// step 4: return the result
	return user, nil
}

func (u *UserRepositoryImpl) InsertViaTnx(ctx context.Context, user *models.User) (string, error) {
	log := u.userRepositoryLog.Method("InsertViaTnx").WithContext(ctx)

	// step 0: begin transaction
	tx := u.db.Begin()
	if tx.Error != nil {
		log.Errorf("Error creating user: %v\n", tx.Error)
		return "", tx.Error
	}

	// step 1: prepare the query
	query := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"

	// step 2: execute the query
	result := tx.Exec(query, user.Name, user.Email, user.Password)

	// step 3: check for errors
	if result.Error != nil {

		tx.Rollback()

		err := pg.HandlePgError(result.Error)
		log.Errorf("Error creating user: %v\n", err)
		return "", err
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {

		tx.Rollback()

		log.Errorf("No user was created.")
		return "", fmt.Errorf("No user was created.")
	}

	// step 5: commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Errorf("Error creating user: %v\n", err)
		return "", err
	}

	// step 6: return the result
	return fmt.Sprintf("Created user (rows affected: %d)\n", rowsAffected), nil
}

func (u *UserRepositoryImpl) InsertViaTnxUsingBatchProcessing(ctx context.Context, users []*models.User) (string, error) {
	log := u.userRepositoryLog.Method("InsertViaTnxUsingBatchProcessing").WithContext(ctx)

	// step 0: begin transaction
	tx := u.db.Begin()
	if tx.Error != nil {
		log.Errorf("Error creating user: %v\n", tx.Error)
		return "", tx.Error
	}

	// step 1: prepare the query
	query := "INSERT INTO users (name, email, password) VALUES "
	values := []interface{}{}
	placeholders := []string{}
	for _, user := range users {
		placeholders = append(placeholders, "(?, ?, ?)")
		values = append(values, user.Name, user.Email, user.Password)
	}
	query += strings.Join(placeholders, ",")

	// step 2: execute the query
	result := tx.Exec(query, values...)

	// step 3: check for errors
	if result.Error != nil {

		tx.Rollback()

		err := pg.HandlePgError(result.Error)
		log.Errorf("Error creating user: %v\n", err)
		return "", err
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		tx.Rollback()
		log.Errorf("No user was created.")
		return "", fmt.Errorf("No user was created.")
	}

	// step 5: commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Errorf("Error creating user: %v\n", err)
		return "", err
	}

	// step 6: return the result
	return fmt.Sprintf("Created user (rows affected: %d)\n", rowsAffected), nil

	/*
		Optimization
			- If using GORM, tx.CreateInBatches(users, 1000)
			- PostgreSQL COPY (FASTEST) : That is 10x–50x faster than INSERT.
				```
				COPY users (name, email, password)
				FROM STDIN WITH CSV
				```
	*/
}
