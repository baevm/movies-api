package users

import (
	"errors"
	"movies-api/internal/validator"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	plaintext *string
	hash      []byte
}

func (p *Password) Set(plainTextPass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPass), 12)

	if err != nil {
		return err
	}

	p.plaintext = &plainTextPass
	p.hash = hash

	return nil
}

func (p *Password) Matches(plainTextPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPass))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "Email must be provided")
	v.Check(validator.Matches(email, validator.EmailRegex), "email", "Email must be valid")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "Password must be provided")
	v.Check(len(password) >= 8, "password", "Password must be greater than 8 characters")
	v.Check(len(password) <= 72, "password", "Password must be greater than 72 characters")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "Name must be provided")
	v.Check(len(user.Name) <= 500, "name", "Name must be less than 500 characters")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
