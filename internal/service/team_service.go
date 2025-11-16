package service

import (
	"context"
	"errors"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/DeadlyParkour777/pr-service/internal/store"
)

type TeamService struct {
	repo TeamRepository
}

func NewTeamService(repo TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) Create(ctx context.Context, team model.Team, members []model.User) (*model.Team, []model.User, error) {
	createdTeam, err := s.repo.AddTeamWithMembers(ctx, team, members)
	if err != nil {
		if errors.Is(err, store.ErrTeamExists) {
			return nil, nil, ErrTeamExists
		}

		return nil, nil, err
	}

	return createdTeam, members, nil
}

func (s *TeamService) Get(ctx context.Context, name string) (*model.Team, []model.User, error) {
	team, members, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, nil, ErrNotFound
		}

		return nil, nil, err
	}

	return team, members, nil
}


