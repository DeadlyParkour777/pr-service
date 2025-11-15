package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPRExists = errors.New("PR with this id already exists")

type PullRequestStore struct {
	conn *pgxpool.Pool
}

func (s *PullRequestStore) Create(ctx context.Context, pr model.PullRequest) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	prQuery := `INSERT INTO pull_requests (id, name, author_id) VALUES ($1, $2, $3);`
	if _, err := tx.Exec(ctx, prQuery, pr.ID, pr.Name, pr.AuthorID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolationCode {
			return ErrPRExists
		}
		return fmt.Errorf("failed to insert PR: %w", err)
	}

	if len(pr.AssignedReviewers) > 0 {
		rows := make([][]any, len(pr.AssignedReviewers))
		for i, reviewerID := range pr.AssignedReviewers {
			rows[i] = []any{pr.ID, reviewerID}
		}

		_, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"pull_request_reviewers"},
			[]string{"pull_request_id", "reviewer_id"},
			pgx.CopyFromRows(rows),
		)

		if err != nil {
			return fmt.Errorf("failed to insert reviewers: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *PullRequestStore) GetByID(ctx context.Context, id string) (*model.PullRequest, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	prQuery := `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	var pr model.PullRequest
	err = tx.QueryRow(ctx, prQuery, id).Scan(
		&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	reviewerQuery := `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1	
	`
	rows, err := tx.Query(ctx, reviewerQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer id: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reviewer rows: %w", err)
	}

	pr.AssignedReviewers = reviewers

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pr, nil
}

func (s *PullRequestStore) Merge(ctx context.Context, id string) error {
	query := `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = NOW()
		WHERE id = $1 AND status = 'OPEN'	
	`

	commandTag, err := s.conn.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		checkQuery := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`
		var exists bool
		if err := s.conn.QueryRow(ctx, checkQuery, id).Scan(&exists); err != nil || !exists {
			return ErrNotFound
		}
	}

	return nil
}

func (s *PullRequestStore) GetByReviewerID(ctx context.Context, reviewerID string) ([]model.PullRequest, error) {
	query := `
		SELECT p.id, p.name, p.author_id, p.status
		FROM pull_requests AS p
		JOIN pull_requests_reviewers AS prr ON p.id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
	`

	rows, err := s.conn.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query PR by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan pr for reviewer: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error prs for reviewer: %w", err)
	}

	return prs, nil
}
