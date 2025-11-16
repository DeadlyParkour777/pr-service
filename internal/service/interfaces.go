package service

import (
	"context"

	"github.com/DeadlyParkour777/pr-service/internal/model"
)

type TeamRepository interface {
	AddTeamWithMembers(ctx context.Context, team model.Team, members []model.User) (*model.Team, error)
	GetByName(ctx context.Context, name string) (*model.Team, []model.User, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*model.FullUserInfo, error)
	SetIsActive(ctx context.Context, id string, isActive bool) (*model.FullUserInfo, error)
	GetActiveTeamMembers(ctx context.Context, teamID int, excludeUserID string) ([]model.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr model.PullRequest) error
	GetByID(ctx context.Context, id string) (*model.PullRequest, error)
	Merge(ctx context.Context, id string) error
	GetByReviewerID(ctx context.Context, reviewerID string) ([]model.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}
