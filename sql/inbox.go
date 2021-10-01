package sql

func (db *sqlImpl) GetInbox(to string) ([]ReceivedMessage, error) {
	var messages []ReceivedMessage
	err := db.db.Select(&messages, "SELECT * FROM receivedmsgs WHERE to_email=$1", to)
	//err := db.db.Select(&messages, "SELECT * FROM recievedmsgs")
	if err != nil {
		return nil, err
	}
	return messages, nil
}
