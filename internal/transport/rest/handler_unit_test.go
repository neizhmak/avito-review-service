package rest

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neizhmak/avito-review-service/internal/service"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "not found",
			err:        &service.ServiceError{Code: service.ErrCodeNotFound, Msg: "missing"},
			wantStatus: http.StatusNotFound,
			wantCode:   service.ErrCodeNotFound,
		},
		{
			name:       "team exists",
			err:        &service.ServiceError{Code: service.ErrCodeTeamExists, Msg: "dup"},
			wantStatus: http.StatusBadRequest,
			wantCode:   service.ErrCodeTeamExists,
		},
		{
			name:       "conflict codes",
			err:        &service.ServiceError{Code: service.ErrCodePRMerged, Msg: "merged"},
			wantStatus: http.StatusConflict,
			wantCode:   service.ErrCodePRMerged,
		},
		{
			name:       "unknown service code",
			err:        &service.ServiceError{Code: "CUSTOM", Msg: "oops"},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "ERROR",
		},
		{
			name:       "non service error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, _ := mapError(tt.err)
			if status != tt.wantStatus || code != tt.wantCode {
				t.Fatalf("want status %d code %s, got %d %s", tt.wantStatus, tt.wantCode, status, code)
			}
		})
	}
}

func TestHandler_ValidationErrors(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		body       string
		query      string
		wantStatus int
	}{
		{
			name:       "createPR missing required",
			handler:    h.createPR,
			body:       `{"pull_request_id":"","pull_request_name":"","author_id":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "mergePR missing id",
			handler:    h.mergePR,
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "reassign missing ids",
			handler:    h.reassignReviewer,
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "getTeam missing query",
			handler:    h.getTeam,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "getUserReviews missing query",
			handler:    h.getUserReviews,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "setUserActive missing id",
			handler:    h.setUserActive,
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/?"+tt.query, bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			tt.handler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
