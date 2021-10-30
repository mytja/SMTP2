package sql

import "fmt"

type Attachment struct {
	ID           int
	MessageID    int    `db:"message_id"`
	OriginalName string `db:"original_name"`
	Filename     string
}

func NewAttachment(ID int, MessageID int, OriginalName string, Filename string) Attachment {
	return Attachment{
		ID:           ID,
		MessageID:    MessageID,
		OriginalName: OriginalName,
		Filename:     Filename,
	}
}

func (db *sqlImpl) CommitAttachment(attachment Attachment) error {
	_, err := db.tx.NamedExec(
		"INSERT INTO attachments (id, message_id, original_name, filename) VALUES (:id, :message_id, :original_name, :filename)",
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
	err := DB.GetDB().Get(&id, "SELECT id FROM attachments WHERE id = (SELECT MAX(id) FROM attachments)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		fmt.Println(err)
		return -1
	}
	return id + 1
}
