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

func (s *PullRequestService) Merge(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	if pr.Status == model.StatusMerged {
		return pr, nil
	}

	err = s.prRepo.Merge(ctx, prID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	mergedPR, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	return mergedPR, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prID, oldReviewerID string) (*model.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, "", ErrNotFound
		}

		return nil, "", err
	}

	if pr.Status == model.StatusMerged {
		return nil, "", ErrPRMerged
	}

	isAssigned := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			isAssigned = true
			break
		}
	}

	if !isAssigned {
		return nil, "", ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, "", ErrNotFound
		}

		return nil, "", err
	}

	allActiveMembers, err := s.userRepo.GetActiveTeamMembers(ctx, oldReviewer.TeamID, "")
	if err != nil {
		return nil, "", err
	}

	forbiddenIDs := make(map[string]struct{})
	forbiddenIDs[pr.AuthorID] = struct{}{}
	for _, reviewer := range pr.AssignedReviewers {
		forbiddenIDs[reviewer] = struct{}{}
	}

	var candidates []model.User
	for _, member := range allActiveMembers {
		if _, isForbidden := forbiddenIDs[member.ID]; !isForbidden {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidates
	}

	s.rnd.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	newReviewer := candidates[0]

	err = s.prRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewer.ID)
	if err != nil {
		return nil, "", err
	}

	updatedPR, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer.ID, nil
}

func (s *PullRequestService) GetByID(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return pr, nil
}
