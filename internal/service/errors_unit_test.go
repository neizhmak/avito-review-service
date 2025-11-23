package service

import "testing"

func TestServiceErrorError(t *testing.T) {
	err := &ServiceError{Code: ErrCodeNotFound, Msg: "not found"}
	if err.Error() != "not found" {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}
