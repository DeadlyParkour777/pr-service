CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    team_id BIGINT NOT NULL,
    CONSTRAINT fk_team
        FOREIGN KEY(team_id)
        REFERENCES teams(id)
        ON DELETE CASCADE
);
CREATE INDEX idx_users_team_id ON users(team_id);

CREATE TABLE IF NOT EXISTS pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    name TEXT NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ,
    CONSTRAINT fk_author
        FOREIGN KEY(author_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    reviewer_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (pull_request_id, reviewer_id),
    CONSTRAINT fk_pr
        FOREIGN KEY(pull_request_id)
        REFERENCES pull_requests(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_reviewer
        FOREIGN KEY(reviewer_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);



