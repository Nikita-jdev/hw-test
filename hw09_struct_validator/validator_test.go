package hw09structvalidator

import (
	"errors"
	"testing"
)

type User struct {
	ID     string   `validate:"len:36"`
	Age    int      `validate:"min:18|max:50"`
	Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
	Role   string   `validate:"in:admin,stuff"`
	Phones []string `validate:"len:11"`
}

func TestValidate(t *testing.T) {
	ok := User{
		ID:     "12345678-1234-1234-1234-123456789012",
		Age:    30,
		Email:  "user@mail.ru",
		Role:   "admin",
		Phones: []string{"89111111111"},
	}
	tests := []struct {
		name   string
		in     interface{}
		wantOK bool
	}{
		{"valid struct", ok, true},
		{"no tags", struct{ A int }{5}, true},
		{"not a struct", 42, false},
		{"bad len", User{ID: "x"}, false},
		{"age out of range", User{ID: ok.ID, Age: 5}, false},
		{"bad email", User{ID: ok.ID, Age: 30, Email: "bad"}, false},
		{"role not in set", User{ID: ok.ID, Age: 30, Email: ok.Email, Role: "guest"}, false},
		{"short phone", User{ID: ok.ID, Age: 30, Email: ok.Email, Role: "admin", Phones: []string{"123"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if (err == nil) != tt.wantOK {
				t.Fatalf("Validate() = %v, wantOK = %v", err, tt.wantOK)
			}
			if err != nil && !errors.Is(err, ErrNotStruct) {
				var errs ValidationErrors
				if !errors.As(err, &errs) {
					t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestValidateErrors(t *testing.T) {
	var errs ValidationErrors
	if err := Validate(User{ID: "x", Age: 5, Email: "bad", Role: "guest"}); !errors.As(err, &errs) {
		t.Fatalf("expected ValidationErrors, got %v", err)
	}
	if len(errs) != 4 {
		t.Fatalf("expected 4 errors, got %d: %v", len(errs), errs)
	}
	for _, e := range errs {
		switch {
		case e.Field == "ID" && !errors.Is(e.Err, ErrLen):
			t.Errorf("ID: want ErrLen, got %v", e.Err)
		case e.Field == "Age" && !errors.Is(e.Err, ErrMin):
			t.Errorf("Age: want ErrMin, got %v", e.Err)
		case e.Field == "Email" && !errors.Is(e.Err, ErrRegexp):
			t.Errorf("Email: want ErrRegexp, got %v", e.Err)
		case e.Field == "Role" && !errors.Is(e.Err, ErrIn):
			t.Errorf("Role: want ErrIn, got %v", e.Err)
		}
	}
}

func TestValidateNotStruct(t *testing.T) {
	if err := Validate("str"); !errors.Is(err, ErrNotStruct) {
		t.Fatalf("expected ErrNotStruct, got %v", err)
	}
}
