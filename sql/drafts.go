package sql

func NewDraft(title string, to string, from string, body string) SentMessage {
	return SentMessage{
		Title:     title,
		ToEmail:   to,
		FromEmail: from,
		Body:      body,
		Pass:      "",
		IsDraft:   true,
	}
}

func (db *sqlImpl) CommitDraftMessage(draft SentMessage) error {
	return nil
}
