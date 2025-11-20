package domain

type User struct {
	ID       int
	Name     string
	isActive bool
	Team     *Team
}

func NewUser(id int, name string) *User {
	return &User{
		ID:       id,
		Name:     name,
		isActive: false,
	}
}

func (u *User) ChangeStatus(status bool) {
	u.isActive = status
}
