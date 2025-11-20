package domain

import (
	"errors"
	"math/rand"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)
const (
	noReviewersError     = "no reviewers"
	noReviewerFoundError = "no reviewer found"
	prIsMergedError      = "pull request is merged"
	noActiveUsersError   = "no active users"
)
const (
	MAX_REVIEWERS = 2
)

type PullRequest struct {
	ID        int
	Title     string
	Author    *User
	Status    PRStatus
	Reviewers []*User
}

func NewPullRequest(id int, title string, author *User) (*PullRequest, error) {
	var reviewers []*User = author.Team.GetRandomActives(author)
	return &PullRequest{
		ID:        id,
		Title:     title,
		Author:    author,
		Status:    PRStatusOpen,
		Reviewers: reviewers,
	}, nil
}

func (pr *PullRequest) ChangeReviewer(personToChange *User) error {
	if pr.Status != PRStatusOpen {
		return errors.New(prIsMergedError)
	}
	for index, reviewer := range pr.Reviewers {
		if reviewer.ID == personToChange.ID {
			activeUsers := personToChange.Team.GetActiveUsers()
			if len(activeUsers) == 0 {
				return errors.New(noActiveUsersError)
			}
			newReviewer := activeUsers[rand.Intn(len(activeUsers))]
			pr.Reviewers[index] = newReviewer
			return nil
		}
	}
	return errors.New(noReviewerFoundError)
}

func (pr *PullRequest) MergePullRequest() {
	pr.Status = PRStatusMerged
}
