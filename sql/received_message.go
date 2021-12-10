package sql

type ReceivedMessage struct {
	ID         int
	Title      string
	URI        string
	ToEmail    string `db:"to_email"`
	FromEmail  string `db:"from_email"`
	ServerID   int    `db:"server_id"`   // This is used to get specific message from server
	ServerPass string `db:"server_pass"` // This is password used to access this email from server
	Warning    string
	MVPPass    string `db:"mvp_pass"`
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
		"INSERT INTO receivedmsgs (id, title, uri, to_email, from_email, server_id, server_pass, warning, mvp_pass) VALUES (:id, :title, :uri, :to_email, :from_email, :server_id, :server_pass, :warning, :mvp_pass)",
		msg)
	if err != nil {
		return err
	}
	err = db.Commit()
	if err != nil {
		return err
	}
	db.logger.Info("Received new email")
	db.logger.Info(res)
	return nil
}

func NewReceivedMessage(
	id int, title string, URI string, to string, from string, sid int, pass string, warning string, mvppass string) ReceivedMessage {
	return ReceivedMessage{
		ID:         id,
		Title:      title,
		URI:        URI,
		ToEmail:    to,
		FromEmail:  from,
		ServerID:   sid,
		ServerPass: pass,
		Warning:    warning,
		MVPPass:    mvppass,
	}
}
