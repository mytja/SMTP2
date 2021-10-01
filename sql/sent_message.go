package sql

import (
	"errors"
	"fmt"
)

type SentMessage struct {
	ID        int `db:"uid"`
	Title     string
	Body      string
	ToEmail   string `db:"to_email"`
	FromEmail string `db:"from_email"`
	Pass      string
}

func (db *sqlImpl) GetSentMessage(id int, pass string) (*SentMessage, error) {
	var message SentMessage

	err := db.db.Get(&message, "SELECT * FROM sentmsgs WHERE uid=$1", id)
	fmt.Println("MSGGETID:", message.ID)
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
		"INSERT INTO sentmsgs (uid, title, body, to_email, from_email, pass) VALUES (:uid, :title, :body, :to_email, :from_email, :pass); SELECT last_insert_rowid();",
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
	err := DB.GetDB().Get(&id, "SELECT uid FROM sentmsgs WHERE uid = (SELECT MAX(uid) FROM sentmsgs)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetLastReceivedID() int {
	var id int
	err := DB.GetDB().Get(&id, "SELECT uid FROM receivedmsgs WHERE uid = (SELECT MAX(uid) FROM receivedmsgs)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}

func NewSentMessage(title string, to string, from string, body string, pass string) SentMessage {
	return SentMessage{
		Title:     title,
		ToEmail:   to,
		FromEmail: from,
		Body:      body,
		Pass:      pass,
	}
}
