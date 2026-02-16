package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(username string, email string, password string) error
	GetByID(id string) (*User, error)
	GetAll() ([]*User, error)
	Update(id string, username string, email string) (string, error)
	SoftDelete(id string) (string, error)
	HardDelete(id string) (string, error)


	// user specific methods
	GetByEmail(email string) (*User, error)
}

type UserRepositoryImpl struct {
	// Add fields for database connection, etc.
	db *gorm.DB
}

func NewUserRepository(_db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{
		db: _db,
	}
}

func (u *UserRepositoryImpl) Create(username string, email string, password string) error {
	fmt.Println("creating user in user repository.")

	// step 0: create a user instance
	// user := &models.User{
	// 	Name:     username,
	// 	Email:    email,
	// 	Password: password,
	// }

	// step 1: prepare the query
	query := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"

	// step 2: execute the query
	result := u.db.Exec(query, username, email, password)

	// step 3: check for errors
	if result.Error != nil {
		// fmt.Printf("Error creating user: %v\n", result.Error)
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf("unique constraint violation")
			case "23503": // foreign_key_violation
				return fmt.Errorf("foreign key violation.")
			case "23502": // not_null_violation
				return fmt.Errorf("not null violation.")
			default:
				return fmt.Errorf("database error: %v", pgErr.Message)
			}
		}
		return result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		fmt.Println("No user was created.")
		return nil
	}

	fmt.Printf("Created user (rows affected: %d)\n",
		rowsAffected)

	// step 5: return the result
	return nil
}

func (u *UserRepositoryImpl) GetByID(id string) (*User, error) {
	fmt.Println("Fetching user by id in user repository.")

	// step 1: prepare the query
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE deleted_at IS NULL AND id = ?"

	// step 2: execute the query
	row := u.db.Raw(query, id).Row()

	// step 3: process the result
	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("User not found.")
			return nil, err
		}
		fmt.Printf("Error fetching user: %v\n", err)
		return nil, err
	}

	// step 4: return the result
	fmt.Printf("Fetched user: %+v\n", user)
	return user, nil
}

func (u *UserRepositoryImpl) GetAll() ([]*User, error) {
	fmt.Println("Fetching all users in user repository.")

	// step 1: prepare the query
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE deleted_at IS NULL"

	// step 2: execute the query
	rows, err := u.db.Raw(query).Rows()
	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
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
	var users []*User
	for rows.Next() {
		var user User
		err := u.db.ScanRows(rows, &user)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			return nil, err
		}
		users = append(users, &user)
	}

	// step 5: return the result
	for _, user := range users {
		fmt.Println(user)
	}
	return users, nil
}

func (u *UserRepositoryImpl) Update(id string, username string, email string) (string, error) {
	fmt.Println("updating user in user repository.")

	// step 1: prepare the query
	query := "UPDATE users SET name = ?, email = ?, updated_at = NOW() WHERE deleted_at IS NULL AND id = ?"

	// step 2: execute the query
	result := u.db.Exec(query, username, email, id)

	// step 3: check for errors
	if result.Error != nil {
		fmt.Printf("Error updating user: %v\n", result.Error)
		return "", result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		fmt.Println("No user was updated.")
		return "", fmt.Errorf("No user was updated.")
	}

	fmt.Printf("Updated user (rows affected: %d)\n",
		rowsAffected)

	// step 5: return the result
	return fmt.Sprintf("User updated successfully (rows affected: %d)", rowsAffected), nil
}

func (u *UserRepositoryImpl) SoftDelete(id string) (string, error) {
	fmt.Println("deleting user in user repository.")

	// step 1: prepare the query
	query := "UPDATE users SET name = ?, email = ?, deleted_at = NOW() WHERE deleted_at IS NULL AND id = ?"

	// step 2: execute the query
	result := u.db.Exec(query, id)

	// step 3: check for errors
	if result.Error != nil {
		fmt.Printf("Error deleting user: %v\n", result.Error)
		return "", result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		fmt.Println("No user was deleted.")
		return "", fmt.Errorf("No user was deleted.")
	}

	fmt.Printf("Deleted user (rows affected: %d)\n", rowsAffected)

	// step 5: return the result
	return fmt.Sprintf("Deleted user (rows affected: %d)\n",rowsAffected), nil
}


func (u *UserRepositoryImpl) HardDelete(id string) (string, error) {
	fmt.Println("deleting user in user repository.")

	// step 1: prepare the query
	query := "DELETE FROM users WHERE id = ?"

	// step 2: execute the query
	result := u.db.Exec(query, id)

	// step 3: check for errors
	if result.Error != nil {
		fmt.Printf("Error deleting user: %v\n", result.Error)
		return "", result.Error
	}

	// step 4: evaluate the result
	rowsAffected := result.RowsAffected
	if rowsAffected == 0 {
		fmt.Println("No user was deleted.")
		return "", fmt.Errorf("No user was deleted.")
	}

	fmt.Printf("Deleted user (rows affected: %d)\n", rowsAffected)

	// step 5: return the result
	return fmt.Sprintf("Deleted user (rows affected: %d)\n",rowsAffected), nil
}


func (u *UserRepositoryImpl) GetByEmail(email string) (*User, error) {
	fmt.Println("Fetching user by email in user repository.")

	// step 1: prepare the query
	query := "SELECT name, email, password FROM users WHERE deleted_at IS NULL AND email = ?"

	// step 2: execute the query
	row := u.db.Raw(query, email).Row()

	// step 3: process the result
	user := &User{}
	err := row.Scan(&user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("User not found.")
			return nil, err
		}
		fmt.Printf("Error fetching user: %v\n", err)
		return nil, err
	}

	// step 4: return the result
	fmt.Printf("Fetched user: %+v\n", user)
	return user, nil
}