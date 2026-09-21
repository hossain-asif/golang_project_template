package identityDTO

import (
	customvalidation "go_project_structure/internal/utils/custom_validation"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type UpdateUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (u UpdateUserRequest) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name,
			validation.Length(3, 255),
			validation.By(customvalidation.NameValidator),
		),
		validation.Field(&u.Email,
			is.Email,
		),
	)
}
