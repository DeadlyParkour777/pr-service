package model

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type Team struct {
	ID   int
	Name string
}

type User struct {
	ID       string
	Username string
	IsActive bool
	TeamID   int
}

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type FullUserInfo struct {
	User
	TeamName string
}
