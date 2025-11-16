package model

type User struct {
	ID       string
	Username string
	IsActive bool
	TeamID   int
}

type FullUserInfo struct {
	User
	TeamName string
}
