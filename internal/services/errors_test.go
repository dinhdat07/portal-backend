package services

import "testing"

func TestAppError_Error(t *testing.T) {
	appErr := &AppError{Code: "x", Message: "hello"}
	if got := appErr.Error(); got != "hello" {
		t.Fatalf("expected message 'hello', got %q", got)
	}
}
