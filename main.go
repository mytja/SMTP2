package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/httphandlers"
	"github.com/mytja/SMTP2/sql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type ServerConfig struct {
	Debug bool
	Host  string
	Port  string
}

func main() {
	fmt.Println("Starting SMTP2 server...")

	config := ServerConfig{}

	command := &cobra.Command{
		Use:   "SMTP2-server",
		Short: "Mail server using SMTP2 protocol",
	}

	command.Flags().BoolVar(&config.Debug, "debug", false, "enable debug mode")
	command.Flags().StringVar(&config.Host, "host", "0.0.0.0", "set server host")
	command.Flags().StringVar(&config.Port, "port", "8080", "set server port")
	command.Flags().StringVar(&constants.ServerUrl, "url", "http://0.0.0.0:8080", "set server URL")
	command.Flags().StringVar(&constants.DbName, "dbname", "smtp2.db", "set DB name")

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	var logger *zap.Logger
	var err error

	if config.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err.Error())
		return
	}

	sugared := logger.Sugar()

	sql.DB, sql.DBERR = sql.NewSQL()
	sql.DB.Init()

	if sql.DBERR != nil {
		sugared.Fatal("Error while creating database: " + sql.DBERR.Error())
		return
	}

	sugared.Info("Database created successfully")

	r := mux.NewRouter()
	r.HandleFunc("/smtp2", httphandlers.WelcomeHandler).Methods(httphandlers.GET)

	// Message & replying
	r.HandleFunc("/smtp2/message/receive", httphandlers.ReceiveMessageHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/new", httphandlers.NewMessageHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/inbox", httphandlers.GetInboxHandler).Methods(httphandlers.GET)
	r.HandleFunc("/smtp2/message/reply/{id}", httphandlers.NewReplyHandler).Methods(httphandlers.POST)
	// Get message from receiver server (ReceivedMessage)
	r.HandleFunc("/smtp2/message/inbox/get/{id}", httphandlers.GetReceivedMessageHandler).Methods(httphandlers.GET)
	// Get message from sender server (SentMessage)
	r.HandleFunc("/smtp2/message/get/{id}", httphandlers.GetSentMessageHandler).Methods(httphandlers.GET)

	// Drafts
	r.HandleFunc("/smtp2/draft/new", httphandlers.NewDraft).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/draft/save", httphandlers.UpdateDraft).Methods(httphandlers.POST)

	// Attachment handling
	r.HandleFunc("/smtp2/attachment/upload/{id}", httphandlers.UploadFile).Methods(httphandlers.POST)

	// User functions
	r.HandleFunc("/user/new", httphandlers.NewUser).Methods(httphandlers.POST)
	r.HandleFunc("/user/login", httphandlers.Login).Methods(httphandlers.POST)

	// SMTP2 Sender Server Verification Protocol
	r.HandleFunc("/smtp2/message/verify", httphandlers.MessageVerificationHandlers).Methods(httphandlers.GET)

	sugared.Info("Serving...")
	serve := config.Host + ":" + config.Port
	sugared.Info("Serving on following URL: " + serve)
	err = http.ListenAndServe(serve, r)
	if err != nil {
		sugared.Fatal(err.Error())
	}

	sugared.Info("Done serving...")
}
