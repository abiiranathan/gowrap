package validation_test

import (
	"errors"
	"testing"

	"github.com/abiiranathan/gowrap/validation"
	"github.com/go-playground/validator/v10"
)

type User struct {
	Name  string `binding:"required,max=5"`
	Age   int    `binding:"required,numeric"`
	Email string `binding:"required,email"`
}

func TestValidateStructs(t *testing.T) {
	t.Parallel()

	h := validation.NewValidator("binding")
	user := User{}

	// Validate with struct pointer
	err := h.Validate(&user)
	if err == nil {
		t.Errorf("Expected validation of empty user to fail, got nil error")
	}

	user.Name = "Abiira Nathan"
	user.Email = "emailcompany.com"

	// Validate with struct, should fail on Name, Age and Email
	err = h.Validate(user)
	if err == nil {
		t.Errorf("Expected validation of user to fail, got nil error")
	}

	// First name still exceeds max characters, should fail
	user.Age = 20
	user.Email = "john.doe@gmail.com"
	err = h.Validate(user)

	if err == nil {
		t.Errorf("Expected struct validation of invalid email to fail, got nil error")
	}

	// Should pass
	user.Name = "AN"
	user.Email = "nabiira@yahoo.com"
	user.Age = 20

	err = h.Validate(user)
	if err != nil {
		t.Error(err)
	}

	// Slice of users
	users := []User{
		{Age: 20, Name: "AN", Email: "an@gmail.com"},
		{Age: 30, Name: "Nat", Email: "nat@gmail.com"},
	}

	err = h.Validate(&users)
	if err != nil {
		t.Errorf("error validating slice of structs: %v", err.Error())
	}

	// Array of invalid users
	usersArray := [2]User{
		{Age: 20, Name: "AN"},
		{Age: 30, Name: "Nat"},
	}

	err = h.Validate(&usersArray) // passed as a pointer
	if err == nil {
		t.Error("expected invalid array of structs to fail validation")
	}

	err = h.Validate(usersArray) // passed as a array
	if err == nil {
		t.Error("expected invalid array of structs to fail validation")
	}

	// Test unsupported types
	err = h.Validate(20)
	if !errors.Is(err, validation.ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}

	num := 40
	err = h.Validate(&num)
	if !errors.Is(err, validation.ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}

}

func TestIsValidEmail(t *testing.T) {
	t.Parallel()

	if validation.IsValidEmail("email") {
		t.Errorf("email is invalid but validated as ok")
	}

	if !validation.IsValidEmail("jd@gmail.com") {
		t.Errorf("valid email validates as invalid email")
	}
}

func TestErrorsAlwaysSlice(t *testing.T) {
	t.Parallel()

	// Array of invalid users
	usersArray := [2]User{
		{Age: 20, Name: "AN"},
		{Age: 30, Name: "Nat"},
	}

	h := validation.NewValidator("binding")
	err := h.Validate(&usersArray)

	if _, ok := err.(validator.ValidationErrors); !ok {
		t.Fatalf("validation returned unknown error type: %T", err.Error())
	}
}

// go test -benchmem -memprofile=mem.pb.gz  -run=^$ -bench ^BenchmarkValidate$ github.com/abiiranathan/eclinicgo/router
// pprof -http=:6060 -alloc_objects mem.pb.gz
func BenchmarkValidate(b *testing.B) {
	b.StopTimer()

	h := validation.NewValidator("binding")
	users := []User{
		{Age: 20, Name: "AN", Email: "an@gmail.com"},
		{Age: 30, Name: "Nat", Email: "nat@gmail.com"},
	}

	for i := 0; i <= b.N; i++ {
		b.StartTimer()
		h.Validate(&users)
		b.StopTimer()
	}
}
