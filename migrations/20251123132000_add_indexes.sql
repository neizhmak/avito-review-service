-- +goose Up
-- SQL section 'Up' is executed when you run 'goose up'

CREATE INDEX IF NOT EXISTS idx_users_team_name ON users (team_name);
CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests (author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests (status);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers (reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pull_request ON pr_reviewers (pull_request_id);

-- +goose Down
-- SQL section 'Down' is executed when you run 'goose down'

DROP INDEX IF EXISTS idx_pr_reviewers_pull_request;
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP INDEX IF EXISTS idx_pr_status;
DROP INDEX IF EXISTS idx_pr_author;
DROP INDEX IF EXISTS idx_users_team_name;
