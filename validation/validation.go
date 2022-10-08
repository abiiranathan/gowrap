package validation

import (
	"errors"
	"net/mail"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var (
	// validation of an unsupported type
	ErrUnsupportedType = errors.New("unsupported type")
)

// Interface for the validator.
// Validates structs, pointers to structs,
// Slices, arrays, pointers to slices & arrays.
//
// Based on https://github.com/go-playground/validator/v10
type Validator interface {
	// Validate obj.
	//
	//If validation fails, it return validator.ValidationErrors
	Validate(obj any) error
	SetTagName(tagName string)
}

type validate struct {
	validator *validator.Validate
}

// returns a new validator for tagName
func NewValidator(tagName string) Validator {
	val := validator.New()
	val.SetTagName(tagName)
	return &validate{validator: val}
}

func (val *validate) SetTagName(tagName string) {
	val.validator.SetTagName(tagName)
}

// Validates structs inside a slice, returns validator.ValidationErrors
// Not that it returns the first encountered error
func (val *validate) validateSlice(value reflect.Value) error {
	count := value.Len()

	for i := 0; i < count; i++ {
		if err := val.validator.Struct(value.Index(i).Interface()); err != nil {
			return err.(validator.ValidationErrors)
		}
	}

	return nil
}

// validates structs, pointers to structs and slices/arrays of structs
func (val *validate) Validate(obj any) error {
	value := reflect.ValueOf(obj)

	var err error
	switch value.Kind() {
	case reflect.Ptr:
		elem := value.Elem()
		switch reflect.ValueOf(elem.Interface()).Kind() {
		case reflect.Struct:
			err = val.validator.Struct(elem.Interface())
		case reflect.Slice, reflect.Array:
			err = val.validateSlice(elem)
		default:
			err = ErrUnsupportedType
		}
	case reflect.Struct:
		err = val.validator.Struct(value.Interface())
	case reflect.Slice, reflect.Array:
		err = val.validateSlice(value)
	default:
		err = ErrUnsupportedType
	}
	return err
}

// validates email using net/email pkg
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
