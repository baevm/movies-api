package validator

import "regexp"

var (
	EmailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

func (v *Validator) Check(ok bool, key, msg string) {
	if !ok {
		v.AddError(key, msg)
	}
}

func AllowedValues[T comparable](value T, permittedValues ...T) bool {
	for i, _ := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}

	return false
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, v := range values {
		uniqueValues[v] = true
	}

	return len(values) == len(uniqueValues)
}
