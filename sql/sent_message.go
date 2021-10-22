package sql

import (
	"errors"
	"fmt"
)

type SentMessage struct {
	ID        int
	Title     string
	Body      string
	ToEmail   string `db:"to_email"`
	FromEmail string `db:"from_email"`
	Pass      string
	ReplyTo   int `db:"reply_to"`
}

func (db *sqlImpl) GetSentMessage(id int, pass string) (*SentMessage, error) {
	var message SentMessage

	err := db.db.Get(&message, "SELECT * FROM sentmsgs WHERE id=$1", id)
	//fmt.Println("MSGGETID:", message.ID)
	if err != nil {
		return nil, err
	}
	if message.Pass == pass {
		return &message, nil
	} else {
		return nil, errors.New("unauthenticated")
	}
}

func (db *sqlImpl) CommitSentMessage(msg SentMessage) (int64, error) {
	res, err := db.tx.NamedExec(
		"INSERT INTO sentmsgs (id, title, body, to_email, from_email, pass, reply_to) VALUES (:id, :title, :body, :to_email, :from_email, :pass, :reply_to); SELECT last_insert_rowid();",
		msg)
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	err = db.Commit()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (db *sqlImpl) GetLastSentID() int {
	var id int
	err := DB.GetDB().Get(&id, "SELECT id FROM sentmsgs WHERE id = (SELECT MAX(id) FROM sentmsgs)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}

func NewSentMessage(title string, to string, from string, body string, pass string, reply_to int) SentMessage {
	return SentMessage{
		Title:     title,
		ToEmail:   to,
		FromEmail: from,
		Body:      body,
		Pass:      pass,
		ReplyTo:   reply_to,
	}
}
