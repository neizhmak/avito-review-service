package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/domain"
	"github.com/neizhmak/avito-review-service/internal/service"
	"github.com/neizhmak/avito-review-service/internal/storage/postgres"
	"github.com/neizhmak/avito-review-service/internal/testutil"
)

func TestHandler_Flow(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamName := "http-team"
	testutil.CleanupTeamData(t, db, teamName)

	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)

	h := NewHandler(service.NewPRService(prStorage, userStorage, teamStorage, db))
	srv := httptest.NewServer(h.InitRouter())
	defer srv.Close()

	client := srv.Client()

	createTeamPayload := `{"team_name":"http-team","members":[{"user_id":"http-author","username":"Author","is_active":true},{"user_id":"http-rev1","username":"Rev1","is_active":true},{"user_id":"http-rev2","username":"Rev2","is_active":true}]}`
	resp, err := client.Post(fmt.Sprintf("%s/team/add", srv.URL), "application/json", bytes.NewBufferString(createTeamPayload))
	if err != nil {
		t.Fatalf("createTeam request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	createPRPayload := `{"pull_request_id":"http-pr","pull_request_name":"Handler Test","author_id":"http-author"}`
	resp, err = client.Post(fmt.Sprintf("%s/pullRequest/create", srv.URL), "application/json", bytes.NewBufferString(createPRPayload))
	if err != nil {
		t.Fatalf("createPR request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for createPR, got %d", resp.StatusCode)
	}

	var createBody struct {
		PR domain.PullRequest `json:"pr"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&createBody); err != nil {
		t.Fatalf("failed to decode createPR response: %v", err)
	}
	_ = resp.Body.Close()

	if len(createBody.PR.Reviewers) != 2 {
		t.Fatalf("expected 2 reviewers assigned, got %v", createBody.PR.Reviewers)
	}
	reviewerID := createBody.PR.Reviewers[0]

	resp, err = client.Get(fmt.Sprintf("%s/users/getReview?user_id=%s", srv.URL, reviewerID))
	if err != nil {
		t.Fatalf("getReview request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for getReview, got %d", resp.StatusCode)
	}
	var reviewsBody struct {
		PullRequests []domain.PullRequestShort `json:"pull_requests"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&reviewsBody); err != nil {
		t.Fatalf("failed to decode reviews response: %v", err)
	}
	_ = resp.Body.Close()
	if len(reviewsBody.PullRequests) == 0 {
		t.Fatalf("expected reviews, got none")
	}

	resp, err = client.Get(fmt.Sprintf("%s/team/get?team_name=%s", srv.URL, teamName))
	if err != nil {
		t.Fatalf("get team failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for getTeam, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	setInactivePayload := `{"user_id":"http-rev1","is_active":false}`
	resp, err = client.Post(fmt.Sprintf("%s/users/setIsActive", srv.URL), "application/json", bytes.NewBufferString(setInactivePayload))
	if err != nil {
		t.Fatalf("setIsActive failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for setIsActive, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = client.Post(fmt.Sprintf("%s/pullRequest/merge", srv.URL), "application/json", bytes.NewBufferString(`{"pull_request_id":"http-pr"}`))
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for merge, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = client.Post(fmt.Sprintf("%s/team/deactivate", srv.URL), "application/json", bytes.NewBufferString(`{"team_name":"http-team"}`))
	if err != nil {
		t.Fatalf("deactivate failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for deactivate, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = client.Get(fmt.Sprintf("%s/health/stats", srv.URL))
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for stats, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	t.Cleanup(func() {
		testutil.CleanupTeamData(t, db, teamName)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM pr_reviewers WHERE pull_request_id = $1", "http-pr")
		_, _ = db.ExecContext(context.Background(), "DELETE FROM pull_requests WHERE id = $1", "http-pr")
	})
}

func TestHandler_ErrorFlows(t *testing.T) {
	db := testutil.OpenTestDB(t)
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)

	h := NewHandler(service.NewPRService(prStorage, userStorage, teamStorage, db))
	srv := httptest.NewServer(h.InitRouter())
	defer srv.Close()

	client := srv.Client()

	resp, err := client.Post(fmt.Sprintf("%s/team/deactivate", srv.URL), "application/json", bytes.NewBufferString(`{"team_name":"missing-team"}`))
	if err != nil {
		t.Fatalf("deactivate missing team failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing team, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = client.Post(fmt.Sprintf("%s/users/setIsActive", srv.URL), "application/json", bytes.NewBufferString(`{"user_id":"missing","is_active":false}`))
	if err != nil {
		t.Fatalf("set inactive missing user failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing user, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = client.Get(fmt.Sprintf("%s/users/getReview?user_id=%s", srv.URL, "missing"))
	if err != nil {
		t.Fatalf("getReview missing user failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing user reviews, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}
