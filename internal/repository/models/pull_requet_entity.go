package models

import "time"

type PullRequest struct {
	ID        string
	Title     string
	AuthorID  string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
