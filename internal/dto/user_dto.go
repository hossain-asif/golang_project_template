package dto

import (
	"encoding/json"
	"go_project_structure/utils/custom_validation"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// register user
type RegisterUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterUserRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required,
			validation.Length(3, 255),
			validation.By(custom_validation.NameValidator),
		),
		validation.Field(&r.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(&r.Password,
			validation.Required,
			validation.Length(8, 20),
			validation.By(custom_validation.PasswordValidator),
		),
	)
}

type RegisterUserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// update user
type UpdateUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (u UpdateUserRequest) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name,
			validation.Length(3, 255),
			validation.By(custom_validation.NameValidator),
		),
		validation.Field(&u.Email,
			is.Email,
		),
	)
}

// login user
type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

func (l LoginUserRequest) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(&l.Password,
			validation.Required,
			validation.Length(8, 20),
			validation.By(custom_validation.PasswordValidator),
		),
	)
}

type LoginUserResponse struct {
	Token string `json:"token"`
}

// csv
type UserCSV struct {
	ID        uint      `csv:"id"`
	Name      string    `csv:"name"`
	Email     string    `csv:"email"`
	CreatedAt time.Time `csv:"created_at"`
	UpdatedAt time.Time `csv:"updated_at"`
}

type UserFromTxt struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *UserFromTxt) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID    interface{} `json:"id"`
		Name  string      `json:"name"`
		Email string      `json:"email"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch v := raw.ID.(type) {
	case float64:
		u.ID = int(v) // JSON number → int
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		u.ID = id // JSON string "1" → int 1
	}

	u.Name = raw.Name
	u.Email = raw.Email
	return nil
}

func (u UserFromTxt) MarshalJSON() ([]byte, error) {
    return json.Marshal(&struct {
        ID    string `json:"id"`   // ✅ force string output
        Name  string `json:"name"`
        Email string `json:"email"`
    }{
        ID:    strconv.Itoa(u.ID), // int → "11"
        Name:  u.Name,
        Email: u.Email,
    })
}