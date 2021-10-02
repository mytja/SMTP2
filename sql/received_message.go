package sql

import (
	"fmt"
)

type ReceivedMessage struct {
	ID         int
	Title      string
	URI        string
	ToEmail    string `db:"to_email"`
	FromEmail  string `db:"from_email"`
	ServerID   int    `db:"server_id"`   // This is used to get specific message from server
	ServerPass string `db:"server_pass"` // This is password used to access this email from server
}

func (db *sqlImpl) GetReceivedMessage(id int) (*ReceivedMessage, error) {
	var message ReceivedMessage
	err := db.db.Get(&message, "SELECT * FROM receivedmsgs WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (db *sqlImpl) CommitReceivedMessages(msg ReceivedMessage) error {
	res, err := db.tx.NamedExec(
		"INSERT INTO receivedmsgs (id, title, uri, to_email, from_email, server_id, server_pass) VALUES (:id, :title, :uri, :to_email, :from_email, :server_id, :server_pass)",
		msg)
	err = db.Commit()
	if err != nil {
		return err
	}
	fmt.Println("Received new email")
	fmt.Println(res)
	return nil
}

func (db *sqlImpl) GetLastReceivedID() int {
	var id int
	err := DB.GetDB().Get(&id, "SELECT id FROM receivedmsgs WHERE id = (SELECT MAX(id) FROM receivedmsgs)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}

func NewReceivedMessage(title string, URI string, to string, from string, id int, pass string) ReceivedMessage {
	return ReceivedMessage{
		Title:      title,
		URI:        URI,
		ToEmail:    to,
		FromEmail:  from,
		ServerID:   id,
		ServerPass: pass,
	}
}
