package identityDTO

import (
	customvalidation "go_project_structure/internal/utils/custom_validation"

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
			validation.By(customvalidation.NameValidator),
		),
		validation.Field(&r.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(&r.Password,
			validation.Required,
			validation.Length(8, 20),
			validation.By(customvalidation.PasswordValidator),
		),
	)
}

type RegisterUserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

