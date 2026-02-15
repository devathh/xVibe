package dtos

type MessageModel struct {
	ID       string `json:"id"`
	ChatID   string `json:"chat_id"`
	AuthorID string `json:"author_id"`
	Body     string `json:"body"`
	SentAt   int64  `json:"sent_at"` // UnixMilli
}

type MessageModels struct {
	Messages []MessageModel `json:"messages"`
	HasMore  bool           `json:"has_more"`
}
