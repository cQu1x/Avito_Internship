package domain

import (
	"errors"
	"math/rand"
)

type Team struct {
	Name  string
	Users []*User
}

func NewTeam(name string, users []*User) *Team {
	return &Team{
		Name:  name,
		Users: users,
	}
}

func (t *Team) AddUser(user *User) error {
	var inThisTeam bool = t.ContainsUser(user)
	if inThisTeam {
		return errors.New(userAlreadyInTeamError)
	}
	if user.Team != nil {
		return errors.New(userInAnotherTeamError)
	}
	t.Users = append(t.Users, user)
	return nil
}

func (t *Team) ContainsUser(user *User) bool {
	for _, teamMember := range t.Users {
		if teamMember.ID == user.ID {
			return true
		}
	}
	return false
}

func (t *Team) GetActiveUsers() []*User {
	var activeUsers []*User
	for _, user := range t.Users {
		if user.isActive {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers
}

func (t *Team) GetRandomActives(AuthorOfPR *User) []*User {
	activeUsers := t.GetActiveUsers()
	for index, users := range activeUsers {
		if users.ID == AuthorOfPR.ID {
			activeUsers = append(activeUsers[:index], activeUsers[index+1:]...)
			break
		}
	}
	if len(activeUsers) <= MAX_REVIEWERS {
		return activeUsers
	}
	var firstReviewerIndex int = rand.Intn(len(activeUsers))
	var secondReviewerIndex int = rand.Intn(len(activeUsers))
	for firstReviewerIndex == secondReviewerIndex {
		secondReviewerIndex = rand.Intn(len(activeUsers))
	}
	return []*User{activeUsers[firstReviewerIndex], activeUsers[secondReviewerIndex]}

}
