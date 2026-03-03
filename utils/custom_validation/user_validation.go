package custom_validation

import (
	"errors"
	"regexp"
)

func NameValidator(value interface{}) error {
	nameRegex := regexp.MustCompile(`^[a-zA-Z ]+$`)

	name, ok := value.(string)
	if !ok {
		return errors.New("invalid name format")
	}

	if !nameRegex.MatchString(name) {
		return errors.New("name can only contain letters and spaces")
	}

	return nil
}

func PasswordValidator(value interface{}) error {
	hasLower   := regexp.MustCompile(`[a-z]`)
	hasUpper   := regexp.MustCompile(`[A-Z]`)
	hasDigit   := regexp.MustCompile(`[0-9]`)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)

	password, ok := value.(string)
	if !ok {
		return errors.New("invalid password format")
	}


	if !hasLower.MatchString(password) ||
	!hasUpper.MatchString(password) ||
	!hasDigit.MatchString(password) ||
	!hasSpecial.MatchString(password) {
		return errors.New("password must contain at least one alphabet, one digit, and one special character")
	}
	

	return nil
}
