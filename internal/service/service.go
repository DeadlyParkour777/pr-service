package service

import "errors"

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrPRExists     = errors.New("pr already exists")
	ErrPRMerged     = errors.New("cannot change merged pr")
	ErrNotAssigned  = errors.New("user is not assigned to this pr")
	ErrNoCandidates = errors.New("no active replacement candidate in team")
	ErrNotFound     = errors.New("resource not found")
)

type Service struct {
	Team  *TeamService
	User  *UserService
	PR    *PullRequestService
	Stats *StatsService
}

type Dependencies struct {
	TeamRepo  TeamRepository
	UserRepo  UserRepository
	PRRepo    PullRequestRepository
	StatsRepo StatsRepository
}

func NewService(d Dependencies) *Service {
	teamService := NewTeamService(d.TeamRepo)
	userService := NewUserService(d.UserRepo, d.PRRepo)
	prService := NewPullRequestService(d.PRRepo, d.UserRepo)
	statsService := NewStatsService(d.StatsRepo)

	service := &Service{
		Team:  teamService,
		User:  userService,
		PR:    prService,
		Stats: statsService,
	}

	return service
}
