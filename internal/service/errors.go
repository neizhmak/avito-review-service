package service

const (
	ErrCodeTeamExists  = "TEAM_EXISTS"
	ErrCodePRExists    = "PR_EXISTS"
	ErrCodePRMerged    = "PR_MERGED"
	ErrCodeNotAssigned = "NOT_ASSIGNED"
	ErrCodeNoCandidate = "NO_CANDIDATE"
	ErrCodeNotFound    = "NOT_FOUND"
)

type ServiceError struct {
	Code string
	Msg  string
}

func (e *ServiceError) Error() string {
	return e.Msg
}

func newServiceError(code string, msg string) *ServiceError {
	return &ServiceError{Code: code, Msg: msg}
}

func notFound(msg string) *ServiceError {
	return newServiceError(ErrCodeNotFound, msg)
}

func conflict(code, msg string) *ServiceError {
	return newServiceError(code, msg)
}
