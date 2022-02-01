package sql

type Attachment struct {
	ID             int
	MessageID      int    `db:"message_id"`
	OriginalName   string `db:"original_name"`
	Filename       string
	AttachmentPass string `db:"attachment_pass"`
	IsForwarded    bool   `db:"is_forwarded"`
	URLToForward   string `db:"url_to_forward"`
}

func NewAttachment(ID int, MessageID int, OriginalName string, Filename string, Pass string, IsForwarded bool, URLToForward string) Attachment {
	return Attachment{
		ID:             ID,
		MessageID:      MessageID,
		OriginalName:   OriginalName,
		Filename:       Filename,
		AttachmentPass: Pass,
		IsForwarded:    IsForwarded,
		URLToForward:   URLToForward,
	}
}

func (db *sqlImpl) CommitAttachment(attachment Attachment) error {
	_, err := db.tx.NamedExec(
		"INSERT INTO attachments (id, message_id, original_name, filename, attachment_pass, is_forwarded, url_to_forward) VALUES (:id, :message_id, :original_name, :filename, :attachment_pass, :is_forwarded, :url_to_forward)",
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
		db.logger.Info(err)
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

func (db *sqlImpl) GetAllAttachments(mid int) ([]Attachment, error) {
	var attachment []Attachment

	err := db.db.Select(&attachment, "SELECT * FROM attachments WHERE message_id=$1", mid)
	if err != nil {
		return make([]Attachment, 0), err
	}
	return attachment, nil
}
