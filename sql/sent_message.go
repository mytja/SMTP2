package sql

type SentMessage struct {
	ID        int
	Title     string
	Body      string
	ToEmail   string `db:"to_email"`
	FromEmail string `db:"from_email"`
	Pass      string
}

func (db *sqlImpl) GetSentMessage(id int) (*SentMessage, error) {
	var message SentMessage

	err := db.db.Get(&message, "SELECT * FROM sentmsgs WHERE id=$1", id)
	//fmt.Println("MSGGETID:", message.ID)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (db *sqlImpl) CommitSentMessage(msg SentMessage) error {
	_, err := db.tx.NamedExec(
		"INSERT INTO sentmsgs (id, title, body, to_email, from_email, pass) VALUES (:id, :title, :body, :to_email, :from_email, :pass)",
		msg)
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
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
