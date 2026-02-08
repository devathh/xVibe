package dtos

type Token struct {
	Access           string `json:"access"`
	Refresh          string `json:"refresh"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
