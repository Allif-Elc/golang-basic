package utility

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		e := &AppError{Err: ErrNotFound, Message: "user not found", Code: 404}
		if got := e.Error(); got != "user not found" {
			t.Errorf("got %q, want %q", got, "user not found")
		}
	})
	t.Run("without message", func(t *testing.T) {
		e := &AppError{Err: ErrNotFound, Code: 404}
		if got := e.Error(); got != "not found" {
			t.Errorf("got %q, want %q", got, "not found")
		}
	})
}

func TestAppError_Unwrap(t *testing.T) {
	e := &AppError{Err: ErrNotFound, Code: 404}
	if !errors.Is(e, ErrNotFound) {
		t.Error("Unwrap should return ErrNotFound")
	}
}

func TestNewAppError(t *testing.T) {
	e := NewAppError(ErrValidation, "bad input", 400)
	if e.Err != ErrValidation || e.Message != "bad input" || e.Code != 400 {
		t.Error("NewAppError fields mismatch")
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name    string
		got     *AppError
		wantErr error
		wantMsg string
		wantCode int
	}{
		{"NotFound", NotFoundError("x"), ErrNotFound, "x", http.StatusNotFound},
		{"Validation", ValidationError("x"), ErrValidation, "x", http.StatusBadRequest},
		{"Unauthorized", UnauthorizedError("x"), ErrUnauthorized, "x", http.StatusUnauthorized},
		{"Forbidden", ForbiddenError("x"), ErrForbidden, "x", http.StatusForbidden},
		{"Conflict", ConflictError("x"), ErrConflict, "x", http.StatusConflict},
		{"Internal", InternalError("x"), ErrInternal, "x", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", tt.got.Message, tt.wantMsg)
			}
			if tt.got.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", tt.got.Code, tt.wantCode)
			}
			if !errors.Is(tt.got, tt.wantErr) {
				t.Errorf("Err should wrap %v", tt.wantErr)
			}
		})
	}
}

func TestGetHTTPStatusForError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 200},
		{"AppError NotFound", NotFoundError("x"), 404},
		{"AppError Validation", ValidationError("x"), 400},
		{"string contains 'not found'", errors.New("user not found"), 404},
		{"string contains 'invalid'", errors.New("invalid input"), 400},
		{"string contains 'required'", errors.New("field required"), 400},
		{"string contains 'must'", errors.New("must be > 0"), 400},
		{"string contains 'too'", errors.New("too long"), 400},
		{"string contains 'potentially dangerous'", errors.New("potentially dangerous content"), 400},
		{"string contains 'unauthorized'", errors.New("unauthorized access"), 401},
		{"string contains 'user id not found'", errors.New("user id not found in context"), 401},
		{"string contains 'forbidden'", errors.New("forbidden action"), 403},
		{"string contains 'permission'", errors.New("permission denied"), 403},
		{"string contains 'already exists'", errors.New("record already exists"), 409},
		{"string contains 'duplicate'", errors.New("duplicate entry"), 409},
		{"default", errors.New("some other error"), 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHTTPStatusForError(tt.err); got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	if !IsNotFoundError(ErrNotFound) {
		t.Error("ErrNotFound should be not found")
	}
	if !IsNotFoundError(errors.New("user not found")) {
		t.Error("'user not found' should be not found")
	}
	if IsNotFoundError(errors.New("another error")) {
		t.Error("should not be not found")
	}
}

func TestIsValidationError(t *testing.T) {
	if !IsValidationError(ErrValidation) {
		t.Error("ErrValidation should be validation error")
	}
	if !IsValidationError(errors.New("invalid value")) {
		t.Error("'invalid' should be validation error")
	}
	if !IsValidationError(errors.New("name required")) {
		t.Error("'required' should be validation error")
	}
	if !IsValidationError(errors.New("must be")) {
		t.Error("'must' should be validation error")
	}
	if !IsValidationError(errors.New("too short")) {
		t.Error("'too' should be validation error")
	}
	if !IsValidationError(errors.New("potentially dangerous")) {
		t.Error("'potentially dangerous' should be validation error")
	}
	if IsValidationError(errors.New("not found")) {
		t.Error("'not found' should not be validation error")
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	if !IsUnauthorizedError(ErrUnauthorized) {
		t.Error("ErrUnauthorized should be unauthorized")
	}
	if !IsUnauthorizedError(errors.New("unauthorized access")) {
		t.Error("'unauthorized' should be unauthorized")
	}
	if IsUnauthorizedError(errors.New("not found")) {
		t.Error("'not found' should not be unauthorized")
	}
}
