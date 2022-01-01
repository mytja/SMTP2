package sql

func NewDraftSentMessage(id int, title string, to string, from string, body string) SentMessage {
	return SentMessage{
		ID:        id,
		Title:     title,
		ToEmail:   to,
		FromEmail: from,
		Body:      body,
		Pass:      "",
	}
}

func NewDraftMessage(id int, originalID int, replypass string, replyid string) Message {
	return Message{
		ID:         id,
		OriginalID: originalID,
		ServerID:   -1,
		ReplyPass:  replypass,
		ReplyID:    replyid,
		Type:       "sent",
		IsDraft:    true,
	}
}

func (db *sqlImpl) UpdateDraftSentMessage(draft SentMessage) error {
	_, err := db.db.NamedExec(
		`UPDATE sentmsgs SET to_email=:to_email, from_email=:from_email, title=:title, body=:body, pass=:pass WHERE id=:id`,
		draft,
	)
	return err
}

func (db *sqlImpl) UpdateDraftMessage(draft Message) error {
	_, err := db.db.NamedExec(
		`UPDATE messages SET original_id=:original_id, server_id=:server_id, reply_pass=:reply_pass, reply_id=:reply_id, is_draft=:is_draft WHERE id=:id`,
		draft,
	)
	return err
}
