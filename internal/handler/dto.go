package handler

import "github.com/DeadlyParkour777/pr-service/internal/model"

type CreateTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequestResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

func ConvertCreateTeamDTOToModels(dto CreateTeamRequest) (model.Team, []model.User) {
	teamModel := model.Team{
		Name: dto.TeamName,
	}

	userModels := make([]model.User, len(dto.Members))
	for i, m := range dto.Members {
		userModels[i] = model.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return teamModel, userModels
}

func ConvertTeamModelsToDTO(team model.Team, members []model.User) TeamResponse {
	dtoMembers := make([]TeamMemberDTO, len(members))
	for i, m := range members {
		dtoMembers[i] = TeamMemberDTO{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return TeamResponse{
		TeamName: team.Name,
		Members:  dtoMembers,
	}
}

func ConvertFullUserModelToDTO(user model.FullUserInfo) UserResponse {
	return UserResponse{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func ConvertPRModelToDTO(pr model.PullRequest) PullRequestResponse {
	return PullRequestResponse{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
	}
}
