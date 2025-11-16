package handler

import (
	"context"

	"github.com/DeadlyParkour777/pr-service/internal/model"
)

type TeamService interface {
	Create(ctx context.Context, team model.Team, members []model.User) (*model.Team, []model.User, error)
	Get(ctx context.Context, name string) (*model.Team, []model.User, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*model.FullUserInfo, error)
	GetReviewsForUser(ctx context.Context, userID string) ([]model.PullRequest, error)
}

type PullRequestService interface {
	Create(ctx context.Context, pr model.PullRequest) (*model.PullRequest, error)
	Merge(ctx context.Context, prID string) (*model.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*model.PullRequest, string, error)
	GetByID(ctx context.Context, prID string) (*model.PullRequest, error)
}

type StatsService interface {
	GetUserStats(ctx context.Context) ([]model.UserStats, error)
}
