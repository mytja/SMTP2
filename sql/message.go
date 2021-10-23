package sql

import (
	"errors"
	"fmt"
	"github.com/mytja/SMTP2/objects"
)

func (db *sqlImpl) GetLastMessageID() int {
	var id int
	err := DB.GetDB().Get(&id, "SELECT id FROM messages WHERE id = (SELECT MAX(id) FROM messages)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) CommitMessage(msg objects.Message) error {
	_, err := db.tx.NamedExec(
		"INSERT INTO messages (id, original_id, server_id, reply_pass, reply_id, type) VALUES (:id, :original_id, :server_id, :reply_pass, :reply_id, :type)",
		msg)
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Tukaj se dogaja hokus pokus, bog ne daj, za kaj sem to naredil...
func (db *sqlImpl) GetOriginalMessageFromReplyTo(ReplyTo int) (*objects.Message, error) {
	var message objects.Message
	err := db.db.Get(&message, "SELECT * FROM messages WHERE id=$1", ReplyTo)
	if err != nil {
		return nil, err
	}
	if message.OriginalID == -1 {
		return &message, nil
	} else {
		err := db.db.Get(&message, "SELECT * FROM messages WHERE id=$1", message.OriginalID)
		if err != nil {
			return nil, err
		}
		if message.OriginalID == -1 {
			return &message, nil
		}
		return nil, errors.New("could not find original message")
	}
}

func (db *sqlImpl) GetOriginalMessageFromOriginalID(OriginalID int) (*objects.Message, error) {
	var message objects.Message
	err := db.db.Get(&message, "SELECT * FROM messages WHERE id=$1", OriginalID)
	if err != nil {
		return nil, err
	}
	if message.OriginalID == -1 {
		return &message, nil
	} else {
		return nil, errors.New("could not find original message")
	}
}

func (db *sqlImpl) GetMessageFromReplyTo(ReplyTo int) (*objects.Message, error) {
	var message objects.Message
	err := db.db.Get(&message, "SELECT * FROM messages WHERE id=$1", ReplyTo)
	return &message, err
}
