package identityDTO

import (
	customvalidation "go_project_structure/internal/utils/custom_validation"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

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
			validation.By(customvalidation.PasswordValidator),
		),
	)
}

type LoginUserResponse struct {
	Token string `json:"token"`
}
