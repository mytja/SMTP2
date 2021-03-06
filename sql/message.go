package sql

import (
	"errors"
)

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

func (db *sqlImpl) GetLastMessageID() int {
	var id int
	err := db.GetDB().Get(&id, "SELECT id FROM messages WHERE id = (SELECT MAX(id) FROM messages)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) CommitMessage(msg Message) error {
	_, err := db.tx.NamedExec(
		"INSERT INTO messages (id, original_id, server_id, reply_pass, reply_id, type, is_draft) VALUES (:id, :original_id, :server_id, :reply_pass, :reply_id, :type, :is_draft)",
		msg)
	if err != nil {
		return err
	}
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Tukaj se dogaja hokus pokus, bog ne daj, za kaj sem to naredil...
func (db *sqlImpl) GetOriginalMessageFromReplyTo(ReplyTo int) (*Message, error) {
	var message Message
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

func (db *sqlImpl) GetOriginalMessageFromOriginalID(OriginalID int) (*Message, error) {
	var message Message
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

func (db *sqlImpl) GetMessageFromReplyTo(ReplyTo int) (*Message, error) {
	var message Message
	err := db.db.Get(&message, "SELECT * FROM messages WHERE id=$1", ReplyTo)
	return &message, err
}

func (db *sqlImpl) GetOriginalFromReplyHeaders(ReplyID string, ReplyPass string) (Message, error) {
	var message Message
	err := db.db.Get(
		&message,
		"SELECT * FROM messages WHERE (original_id=$1 AND reply_id=$2 AND reply_pass=$3)",
		-1, ReplyID, ReplyPass,
	)
	return message, err
}

func (db *sqlImpl) DeleteMessage(ID int) error {
	db.GenerateNewTransaction()
	db.tx.MustExec("DELETE FROM messages WHERE id=$1", ID)
	err := db.Commit()
	return err
}

func (db *sqlImpl) IsMessageInDatabase(mID int) (bool, *Message, error) {
	var message Message
	err := db.db.Get(&message, "SELECT * FROM messages WHERE id=$1", mID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &message, nil
}
