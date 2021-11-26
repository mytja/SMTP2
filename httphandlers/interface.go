package httphandlers

import (
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/sql"
	"go.uber.org/zap"
	"net/http"
)

type httpImpl struct {
	logger *zap.SugaredLogger
	db     sql.SQL
	config helpers.ServerConfig
}

type HTTP interface {
	// httpattachment.go
	UploadFile(w http.ResponseWriter, r *http.Request)
	DeleteAttachment(w http.ResponseWriter, r *http.Request)
	GetAttachment(w http.ResponseWriter, r *http.Request)
	RetrieveAttachment(w http.ResponseWriter, r *http.Request)

	// httpmessage.go
	GetReceivedMessageHandler(w http.ResponseWriter, r *http.Request)
	GetSentMessageHandler(w http.ResponseWriter, r *http.Request)
	GetInboxHandler(w http.ResponseWriter, r *http.Request)
	UpdateMessage(w http.ResponseWriter, r *http.Request)
	DeleteMessage(w http.ResponseWriter, r *http.Request)
	RetrieveMessageFromRemoteServer(w http.ResponseWriter, r *http.Request)

	// sendmessage.go
	NewMessageHandler(w http.ResponseWriter, r *http.Request)

	// verification.go
	MessageVerificationHandlers(w http.ResponseWriter, r *http.Request)

	// httpuser.go
	Login(w http.ResponseWriter, r *http.Request)
	NewUser(w http.ResponseWriter, r *http.Request)

	// receivemessage.go
	ReceiveMessageHandler(w http.ResponseWriter, r *http.Request)

	// reply.go
	NewReplyHandler(w http.ResponseWriter, r *http.Request)

	// httpdraft.go
	NewDraft(w http.ResponseWriter, r *http.Request)

	ServerInfo(w http.ResponseWriter, r *http.Request)
}

func NewHTTPInterface(logger *zap.SugaredLogger, db sql.SQL, config helpers.ServerConfig) HTTP {
	return &httpImpl{
		logger: logger,
		db:     db,
		config: config,
	}
}
