package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Store struct {
	conn *pgxpool.Pool
	team *TeamStore
	user *UserStore
	pr   *PullRequestStore
}

func NewStore(databaseURL string) (*Store, error) {
	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Store{conn: conn}, nil
}

func (s *Store) Close() {
	s.conn.Close()
}

func (s *Store) Team() *TeamStore {
	if s.team == nil {
		s.team = &TeamStore{conn: s.conn}
	}

	return s.team
}

func (s *Store) User() *UserStore {
	if s.user == nil {
		s.user = &UserStore{conn: s.conn}
	}

	return s.user
}

func (s *Store) PR() *PullRequestStore {
	if s.pr == nil {
		s.pr = &PullRequestStore{conn: s.conn}
	}

	return s.pr
}
