package dtos

type CreateRequest struct {
	ChatID string `json:"chat_id"`
	Body   string `json:"body"`
}

type DeleteRequest struct {
	MsgID string `json:"msg_id"`
}

type GetRequest struct {
	ChatID   string `json:"chat_id"`
	Limit    uint32 `json:"limit"`
	BeforeID string `json:"before_id"`
}
