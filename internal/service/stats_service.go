package service

import (
	"context"
	"sort"

	"github.com/DeadlyParkour777/pr-service/internal/model"
)

type StatsService struct {
	repo StatsRepository
}

func NewStatsService(repo StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetUserStats(ctx context.Context) ([]model.UserStats, error) {
	counts, err := s.repo.GetReviewCountsByUser(ctx)
	if err != nil {
		return nil, err
	}

	stats := make([]model.UserStats, 0, len(counts))
	for userID, count := range counts {
		stats = append(stats, model.UserStats{UserID: userID, ReviewCount: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].UserID < stats[j].UserID
	})

	return stats, nil
}
