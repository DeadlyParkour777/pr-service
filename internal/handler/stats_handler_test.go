package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"testing"

	"github.com/DeadlyParkour777/pr-service/internal/model"
	"github.com/stretchr/testify/require"
)

func TestGetUserStats(t *testing.T) {
	ctx := context.Background()

	t.Run("success - returns user stats", func(t *testing.T) {
		truncateTables(ctx)

		team := model.Team{Name: "stats-team"}
		users := []model.User{
			{ID: "user1", Username: "User One", IsActive: true},
			{ID: "user2", Username: "User Two", IsActive: true},
			{ID: "user3", Username: "User Three", IsActive: true},
			{ID: "author", Username: "Author", IsActive: true},
		}
		_, err := testStore.Team().AddTeamWithMembers(ctx, team, users)
		require.NoError(t, err)

		prs := []model.PullRequest{
			{ID: "pr1", Name: "PR One", AuthorID: "author", AssignedReviewers: []string{"user1", "user2"}},
			{ID: "pr2", Name: "PR Two", AuthorID: "author", AssignedReviewers: []string{"user1", "user3"}},
			{ID: "pr3", Name: "PR Three", AuthorID: "author", AssignedReviewers: []string{"user2"}},
		}

		for _, pr := range prs {
			err := testStore.PR().Create(ctx, pr)
			require.NoError(t, err)
		}

		token := getTestToken(t, "test-user")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServerURL+"/stats/user", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result struct {
			UserStats []model.UserStats `json:"user_stats"`
		}
		err = json.Unmarshal(bodyBytes, &result)
		require.NoError(t, err)

		sort.Slice(result.UserStats, func(i, j int) bool {
			return result.UserStats[i].UserID < result.UserStats[j].UserID
		})

		expectedStats := []model.UserStats{
			{UserID: "user1", ReviewCount: 2},
			{UserID: "user2", ReviewCount: 2},
			{UserID: "user3", ReviewCount: 1},
		}

		require.Equal(t, expectedStats, result.UserStats)
	})

	t.Run("success - returns empty list when no reviews exist", func(t *testing.T) {
		truncateTables(ctx)

		token := getTestToken(t, "test-user")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServerURL+"/stats/user", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result struct {
			UserStats []model.UserStats `json:"user_stats"`
		}
		err = json.Unmarshal(bodyBytes, &result)
		require.NoError(t, err)

		require.NotNil(t, result.UserStats)
		require.Len(t, result.UserStats, 0)
	})
}
