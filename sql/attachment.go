package sql

import "fmt"

type Attachment struct {
	ID             int
	MessageID      int    `db:"message_id"`
	OriginalName   string `db:"original_name"`
	Filename       string
	AttachmentPass string `db:"attachment_pass"`
	Type           string
}

func NewAttachment(ID int, MessageID int, OriginalName string, Filename string, Pass string, Type string) Attachment {
	return Attachment{
		ID:             ID,
		MessageID:      MessageID,
		OriginalName:   OriginalName,
		Filename:       Filename,
		AttachmentPass: Pass,
		Type:           Type,
	}
}

func (db *sqlImpl) CommitAttachment(attachment Attachment) error {
	_, err := db.tx.NamedExec(
		"INSERT INTO attachments (id, message_id, original_name, filename, attachment_pass, type) VALUES (:id, :message_id, :original_name, :filename, :attachment_pass, :type)",
		attachment)
	if err != nil {
		return err
	}
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *sqlImpl) GetLastAttachmentID() int {
	var id int
	err := db.db.Get(&id, "SELECT id FROM attachments WHERE id = (SELECT MAX(id) FROM attachments)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetAttachment(mid int, aid int) (*Attachment, error) {
	var attachment Attachment

	err := db.db.Get(&attachment, "SELECT * FROM attachments WHERE message_id=$1 AND id=$2", mid, aid)
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (db *sqlImpl) DeleteAttachment(mid int, aid int) error {
	_, err := db.db.Exec("DELETE FROM attachments WHERE message_id=$1 AND id=$2", mid, aid)
	return err
}
