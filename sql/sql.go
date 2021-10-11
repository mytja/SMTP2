package sql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/objects"
)

type sqlImpl struct {
	db  *sqlx.DB
	tx  *sqlx.Tx
	err error
}

type SQL interface {
	Init()
	GetReceivedMessage(int) (*ReceivedMessage, error)
	GetDB() *sqlx.DB
	CommitReceivedMessages(ReceivedMessage) error
	GetInbox(string) ([]ReceivedMessage, error)
	GenerateNewTransaction()
	NewUser(string, string) error
	GetUserByEmail(string) (objects.User, error)
	Commit() error
	CommitSentMessage(SentMessage) (int64, error)
	GetSentMessage(int, string) (*SentMessage, error)
	GetLastSentID() int
	GetLastReceivedID() int
}

func NewSQL() (SQL, error) {
	db, err := sqlx.Connect("sqlite3", constants.DbName)
	tx := db.MustBegin()
	return &sqlImpl{
		db:  db,
		tx:  tx,
		err: err,
	}, err
}
