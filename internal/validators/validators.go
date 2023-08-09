package validators

import (
	"regexp"
)

var (
	EmailRx = regexp.MustCompile(
		"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validators struct {
	Errors map[string]string
}

func (v *Validators) AddErr(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

func (v *Validators) IsValid() bool {
	return len(v.Errors) == 0
}

func (v *Validators) Check(ok bool, field, message string) bool {
	if !ok {
		v.AddErr(field, message)
	}
	return true
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

func New() *Validators {
	return &Validators{Errors: map[string]string{}}
}

func PermittedValues[T comparable](key T, list ...T) bool {
	for i := range list {
		if key == list[i] {
			return true
		}
	}

	return false
}

// Matches returns true if a string value matches a specific regexp pattern.
func Matches(value string) bool {
	return EmailRx.MatchString(value)
}
