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

const postgresUniqueViolationCode = "23505"

var ErrTeamExists = errors.New("team with this name already exists")

type TeamStore struct {
	conn *pgxpool.Pool
}

func (s *TeamStore) AddTeamWithMembers(ctx context.Context, team model.Team, members []model.User) (*model.Team, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	createTeamQuery := `INSERT INTO teams (name) VALUES ($1) RETURNING id;`
	var teamID int
	err = tx.QueryRow(ctx, createTeamQuery, team.Name).Scan(&teamID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolationCode {
			return nil, ErrTeamExists
		}
		return nil, fmt.Errorf("failed to insert team: %w", err)
	}

	if len(members) > 0 {
		rows := make([][]any, len(members))
		for i, member := range members {
			rows[i] = []any{member.ID, member.Username, member.IsActive, teamID}
		}

		_, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"users"},
			[]string{"id", "username", "is_active", "team_id"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert users: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdTeam := team
	createdTeam.ID = teamID
	return &createdTeam, nil
}

func (s *TeamStore) GetByName(ctx context.Context, name string) (*model.Team, []model.User, error) {
	query := `
		SELECT t.id, t.name, u.id, u.username, u.is_active, u.team_id
		FROM teams AS t
		LEFT JOIN users AS u ON t.id = u.team_id
		WHERE t.name = $1;
	`
	rows, err := s.conn.Query(ctx, query, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query team by name: %w", err)
	}
	defer rows.Close()

	var team model.Team
	var members []model.User
	var teamFound bool

	for rows.Next() {
		var member model.User
		var UserID, username *string
		var isActive *bool
		var teamID *int

		if err := rows.Scan(&team.ID, &team.Name, &UserID, &username, &isActive, &teamID); err != nil {
			return nil, nil, fmt.Errorf("failed to scan team row: %w", err)
		}
		teamFound = true

		if UserID != nil {
			member.ID = *UserID
			member.Username = *username
			member.IsActive = *isActive
			member.TeamID = *teamID
			members = append(members, member)
		}
	}

	if !teamFound {
		return nil, nil, ErrNotFound
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error team rows: %w", err)
	}

	return &team, members, nil
}
