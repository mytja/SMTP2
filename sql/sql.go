package sql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type sqlImpl struct {
	db     *sqlx.DB
	tx     *sqlx.Tx
	err    error
	logger *zap.SugaredLogger
}

type SQL interface {
	Init()
	GetReceivedMessage(int) (*ReceivedMessage, error)
	GetDB() *sqlx.DB
	CommitReceivedMessages(ReceivedMessage) error
	GetInbox(string) ([]ReceivedMessage, error)
	GenerateNewTransaction()
	NewUser(string, string) error
	GetUserByEmail(string) (User, error)
	Commit() error
	CommitSentMessage(SentMessage) error
	GetSentMessage(int) (*SentMessage, error)
	GetLastMessageID() int
	CommitMessage(message Message) error
	GetOriginalMessageFromOriginalID(int) (*Message, error)
	GetOriginalMessageFromReplyTo(int) (*Message, error)
	GetMessageFromReplyTo(int) (*Message, error)
	GetOriginalFromReplyHeaders(string, string) (Message, error)
	DeleteMessage(int) error
	DeleteSentMessage(int) error
	UpdateDraftSentMessage(SentMessage) error
	UpdateDraftMessage(Message) error
	CommitAttachment(Attachment) error
	GetLastAttachmentID() int
	GetAttachment(int, int) (*Attachment, error)
	DeleteAttachment(int, int) error
	GetAllAttachments(int) ([]Attachment, error)
	UpdateReceivedMessage(ReceivedMessage) error
}

func NewSQL(driver string, drivername string, logger *zap.SugaredLogger) (SQL, error) {
	db, err := sqlx.Connect(driver, drivername)
	tx := db.MustBegin()
	return &sqlImpl{
		db:     db,
		tx:     tx,
		err:    err,
		logger: logger,
	}, err
}
