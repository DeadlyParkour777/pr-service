package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/DeadlyParkour777/pr-service/internal/store"
)

type PullRequestService struct {
	prRepo   PullRequestRepository
	userRepo UserRepository
	rnd      *rand.Rand
}

func NewPullRequestService(prRepo PullRequestRepository, userRepo UserRepository) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PullRequestService) Create(ctx context.Context, pr model.PullRequest) (*model.PullRequest, error) {
	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, author.TeamID, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	s.rnd.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	var reviewers []string
	limit := 2
	if len(candidates) < limit {
		limit = len(candidates)
	}

	for _, candidate := range candidates[:limit] {
		reviewers = append(reviewers, candidate.ID)
	}

	pr.AssignedReviewers = reviewers

	if err := s.prRepo.Create(ctx, pr); err != nil {
		if errors.Is(err, store.ErrPRExists) {
			return nil, ErrPRExists
		}

		return nil, err
	}

	prs, err := s.prRepo.GetByID(ctx, pr.ID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}
