package objects

type Message struct {
	ID         int
	OriginalID int    `db:"original_id"`
	ServerID   int    `db:"server_id"`
	ReplyPass  string `db:"reply_pass"`
	ReplyID    string `db:"reply_id"`
	Type       string
	IsDraft    bool `db:"is_draft"`
}

func NewMessage(ID int, OriginalID int, ServerID int, ReplyPass string, ReplyID string, Type string, IsDraft bool) Message {
	return Message{
		ID:         ID,
		OriginalID: OriginalID,
		ServerID:   ServerID,
		ReplyPass:  ReplyPass,
		ReplyID:    ReplyID,
		Type:       Type,
		IsDraft:    IsDraft,
	}
}
