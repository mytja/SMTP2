package sql

import (
	"github.com/mytja/SMTP2/objects"
)

func (db *sqlImpl) NewUser(email string, password string) error {
	user := objects.NewUser(email, password)
	_, err := db.tx.NamedExec("INSERT INTO users (email, pass) VALUES (:email, :pass)", user)
	if err != nil {
		return err
	}
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
}
