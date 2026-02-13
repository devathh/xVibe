package dtos

type CreateRequest struct {
	TypeID int `json:"type_id"`

	CreateSelf *struct {
		MemberID string `json:"member_id"`
	} `json:"create_self"`

	CreateGroup *struct {
		Title     string   `json:"title"`
		MemberIds []string `json:"member_ids"`
	} `json:"create_group"`
}

type DeleteRequest struct {
	ChatID string `json:"chat_id"`
}

type UpdateRequest struct {
	ChatID string `json:"chat_id"`
	Title  string `json:"title"`
}

type MembersRequest struct {
	ChatID    string   `json:"chat_id"`
	MemberIds []string `json:"member_ids"`
}

type GetRequest struct {
	ChatID string `json:"chat_id"`
}
