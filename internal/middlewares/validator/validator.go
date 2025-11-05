package validator

import "github.com/go-playground/validator/v10"

type CustomValid struct {
	validator *validator.Validate
}

func NewValidator() *CustomValid {
	v := validator.New()

	return &CustomValid{v}
}

func (cv *CustomValid) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
