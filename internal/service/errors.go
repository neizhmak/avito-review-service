package service

import "net/http"

const (
	ErrCodeTeamExists  = "TEAM_EXISTS"
	ErrCodePRExists    = "PR_EXISTS"
	ErrCodePRMerged    = "PR_MERGED"
	ErrCodeNotAssigned = "NOT_ASSIGNED"
	ErrCodeNoCandidate = "NO_CANDIDATE"
	ErrCodeNotFound    = "NOT_FOUND"
)

type ServiceError struct {
	Code   string
	Msg    string
	Status int
}

func (e *ServiceError) Error() string {
	return e.Msg
}

func newServiceError(code string, msg string, status int) *ServiceError {
	return &ServiceError{Code: code, Msg: msg, Status: status}
}

func notFound(msg string) *ServiceError {
	return newServiceError(ErrCodeNotFound, msg, http.StatusNotFound)
}

func conflict(code, msg string) *ServiceError {
	return newServiceError(code, msg, http.StatusConflict)
}
