package sql

func (db *sqlImpl) GetInbox(to string) ([]ReceivedMessage, error) {
	var messages []ReceivedMessage
	err := db.db.Select(&messages, "SELECT * FROM receivedmsgs WHERE to_email=$1", to)
	return messages, err
}

func (db *sqlImpl) GetAllSentMessages(from string) ([]SentMessage, error) {
	var messages []SentMessage
	err := db.db.Select(&messages, "SELECT * FROM sentmsgs WHERE from_email=$1", from)
	return messages, err
}
