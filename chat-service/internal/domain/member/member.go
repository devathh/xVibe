package member

import "github.com/google/uuid"

type Member struct {
	id        uuid.UUID
	firstname string
	lastname  string
	isOwner   bool
}

func New(
	id uuid.UUID,
	firstname, lastname string,
	isOwner bool,
) *Member {
	return &Member{
		id:        id,
		firstname: firstname,
		lastname:  lastname,
		isOwner:   isOwner,
	}
}

func (m *Member) ID() uuid.UUID {
	return m.id
}

func (m *Member) Firstname() string {
	return m.firstname
}

func (m *Member) Lastname() string {
	return m.lastname
}

func (m *Member) IsOwner() bool {
	return m.isOwner
}
