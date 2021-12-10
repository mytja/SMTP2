package sql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mytja/SMTP2/objects"
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
	GetUserByEmail(string) (objects.User, error)
	Commit() error
	CommitSentMessage(SentMessage) error
	GetSentMessage(int) (*SentMessage, error)
	GetLastMessageID() int
	CommitMessage(message objects.Message) error
	GetOriginalMessageFromOriginalID(int) (*objects.Message, error)
	GetOriginalMessageFromReplyTo(int) (*objects.Message, error)
	GetMessageFromReplyTo(int) (*objects.Message, error)
	GetOriginalFromReplyHeaders(string, string) (objects.Message, error)
	DeleteMessage(int) error
	DeleteSentMessage(int) error
	UpdateDraftSentMessage(SentMessage) error
	UpdateDraftMessage(objects.Message) error
	CommitAttachment(Attachment) error
	GetLastAttachmentID() int
	GetAttachment(int, int) (*Attachment, error)
	DeleteAttachment(int, int) error
	GetAllAttachments(int) ([]Attachment, error)
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
