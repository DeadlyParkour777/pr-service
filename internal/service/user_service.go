package service

import (
	"context"
	"errors"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/DeadlyParkour777/pr-service/internal/store"
)

type UserService struct {
	userRepo UserRepository
	prRepo   PullRequestRepository
}

func NewUserService(userRepo UserRepository, prRepo PullRequestRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.FullUserInfo, error) {
	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return user, nil
}

func (s *UserService) GetReviewsForUser(ctx context.Context, userID string) ([]model.PullRequest, error) {
	prs, err := s.prRepo.GetByReviewerID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return prs, nil
}
