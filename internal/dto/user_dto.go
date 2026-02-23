package dto

import "time"

type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type RegisterUserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserRequest struct {
	Name  *string `json:"username"`
	Email *string `json:"email"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type LoginUserResponse struct {
	Token string `json:"token"`
}

type UserCSV struct {
	ID        uint      `csv:"id"`
	Name      string    `csv:"name"`
	Email     string    `csv:"email"`
	CreatedAt time.Time `csv:"created_at"`
	UpdatedAt time.Time `csv:"updated_at"`
}
