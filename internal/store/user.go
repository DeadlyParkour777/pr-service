package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	conn *pgxpool.Pool
}

func (s *UserStore) GetByID(ctx context.Context, id string) (*model.FullUserInfo, error) {
	query := `
		SELECT u.id, u.username, u.is_active, u.team_id, t.name AS team_name
		FROM users AS u
		JOIN teams AS t ON u.team_id = t.id
		WHERE u.id = $1;
	`

	var user model.FullUserInfo
	err := s.conn.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.IsActive, &user.TeamID, &user.TeamName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (s *UserStore) SetIsActive(ctx context.Context, id string, isActive bool) (*model.FullUserInfo, error) {
	query := `
		WITH updated_user AS (
			UPDATE users SET is_active = $2 WHERE id = $1
			RETURNING id, username, is_active, team_id
		)
		SELECT u.id, u.username, u.is_active, u.team_id, t.name as team_name
		FROM updated_user AS u
		JOIN teams AS t ON u.team_id = t.id;
	`

	var user model.FullUserInfo
	err := s.conn.QueryRow(ctx, query, id, isActive).Scan(
		&user.ID, &user.Username, &user.IsActive, &user.TeamID, &user.TeamName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to set user active status: %w", err)
	}

	return &user, nil
}

func (s *UserStore) GetActiveTeamMembers(ctx context.Context, teamID int, excludeUserId string) ([]model.User, error) {
	query := `
		SELECT id, username, is_active, team_id
		FROM users
		WHERE team_id = $1 AND is_active = true AND id != $2;	
	`

	rows, err := s.conn.Query(ctx, query, teamID, excludeUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to query active team members: %w", err)
	}
	defer rows.Close()

	var members []model.User
	for rows.Next() {
		var member model.User
		if err := rows.Scan(&member.ID, &member.Username, &member.IsActive, &member.TeamID); err != nil {
			return nil, fmt.Errorf("failed to scan active member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after active members: %w", err)
	}

	return members, nil
}
