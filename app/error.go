package app

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Extracts the tag from the validator field errors.
func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "dir":
		return "Not a valid directory"
	case "file":
		return "Not a valid file path"
	case "max":
		return "Too long"
	case "min":
		return "Too short"
	case "unique":
		return "This field must be unique"
	case "uuid4":
		return "Invalid uuid4"
	case "uuid":
		return "Invalid uuid"
	case "jwt":
		return "Invalid jwt"
	case "json":
		return "Invalid json"
	case "datetime":
		return "Invalid datetime format"
	case "uppercase":
		return "Must be uppercase"
	case "lowercase":
		return "Must be lowercase"
	case "boolean":
		return "Must be a boolean"
	case "alphanum":
		return "Must be a alpha-numeric"
	case "alpha":
		return "Must be alpha characters only"
	case "ascii":
		return "Must be alpha ascii only"
	case "numeric":
		return "Must be a number"
	case "iscolor":
		return "Must be a a valid 'hexcolor|rgb|rgba|hsl|hsla'"
	}

	return fe.Error()
}

// Struct to hold the error message
type validationError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

func (h *Handler) SendError(ctx *fiber.Ctx, err error) {
	var valError validator.ValidationErrors

	if errors.As(err, &valError) {
		out := make([]validationError, len(valError))
		for i, fe := range valError {
			out[i] = validationError{fe.Field(), msgForTag(fe)}
		}

		ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"errors": out})
		return
	}

	ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
}
