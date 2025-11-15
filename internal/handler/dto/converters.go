package dto

import "github.com/DeadlyParkour777/pr-service/internal/model"

func ConvertTeamModelToDTO(team model.Team, members []model.User) Team {
	dtoMember := make([]TeamMember, len(members))
	for i, m := range members {
		dtoMember[i] = TeamMember{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return Team{
		TeamName: team.Name,
		Members:  dtoMember,
	}
}

func ConvertCreateTeamDTOToModel(dto CreateTeamRequest) (model.Team, []model.User) {
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
