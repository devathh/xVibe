package user

import "net/mail"

type Email string

func (e Email) IsValid() bool {
	if _, err := mail.ParseAddress(string(e)); err != nil {
		return false
	}

	return true
}

func (e Email) Value() string {
	return string(e)
}
