package dtos

type ChatModel struct {
	ID         string `json:"id"`
	OwnerID    string `json:"owner_id"`
	Title      string `json:"title"`
	TypeID     int    `json:"type_id"`
	TypeString string `json:"type_string"`
	CreatedAt  int64  `json:"created_at"` // UnixMilli
}

type ChatModels struct {
	Chats     []ChatModel `json:"chats"`
	Timestamp int64       `json:"timestamp"`
}

type ChatWithMembers struct {
	ChatModel      ChatModel `json:"chat_model"`
	Members        []Member  `json:"members"`
	IsCurrentOwner bool      `json:"is_current_owner"`
	Timestamp      int64     `json:"timestamp"`
}

type Member struct {
	UserID    string `json:"user_id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	IsOwner   bool   `json:"is_owner"`
}
