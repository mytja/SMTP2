package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/httphandlers"
	"github.com/mytja/SMTP2/sql"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting SMTP2 server...")

	config := helpers.ServerConfig{}

	command := &cobra.Command{
		Use:   "SMTP2-server",
		Short: "Mail server using SMTP2 protocol",
	}

	command.Flags().BoolVar(&config.Debug, "debug", false, "enable debug mode")
	command.Flags().StringVar(&config.Host, "host", "0.0.0.0", "set server host")
	command.Flags().StringVar(&config.Port, "port", "8080", "set server port")
	command.Flags().StringVar(&constants.ServerUrl, "url", "http://0.0.0.0:8080", "set server URL")
	command.Flags().StringVar(&constants.DbName, "dbname", "smtp2.db", "set DB name")
	command.Flags().StringVar(&config.HostURL, "hosturl", "", "What should be shown after @ symbol")
	command.Flags().BoolVar(&config.HTTPSEnabled, "https", false, "Is https enabled for following domain")

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	if config.HostURL == "" {
		// This means it's running in localhost
		config.HostURL = config.Host + ":" + config.Port
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

	db, err := sql.NewSQL()
	db.Init()

	if err != nil {
		sugared.Fatal("Error while creating database: " + err.Error())
		return
	}

	httphandler := httphandlers.NewHTTPInterface(sugared, db, config)

	sugared.Info("Database created successfully")

	r := mux.NewRouter()
	r.HandleFunc("/smtp2", httphandlers.WelcomeHandler).Methods(httphandlers.GET)

	r.HandleFunc("/smtp2/server/info", httphandler.ServerInfo).Methods(httphandlers.GET)

	// Message & replying
	r.HandleFunc("/smtp2/message/receive", httphandler.ReceiveMessageHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/new", httphandler.NewMessageHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/inbox", httphandler.GetInboxHandler).Methods(httphandlers.GET)
	r.HandleFunc("/smtp2/message/update", httphandler.UpdateMessage).Methods(httphandlers.PATCH)
	r.HandleFunc("/smtp2/message/reply/{id}", httphandler.NewReplyHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/delete/{id}", httphandler.DeleteMessage).Methods(httphandlers.DELETE)
	// Get message from receiver server (ReceivedMessage)
	r.HandleFunc("/smtp2/message/inbox/get/{id}", httphandler.GetReceivedMessageHandler).Methods(httphandlers.GET)
	// Get message from sender server (SentMessage)
	r.HandleFunc("/smtp2/message/get/{id}", httphandler.GetSentMessageHandler).Methods(httphandlers.GET)

	// Drafts
	r.HandleFunc("/smtp2/draft/new", httphandler.NewDraft).Methods(httphandlers.POST)

	// Attachment handling
	r.HandleFunc("/smtp2/attachment/upload/{id}", httphandler.UploadFile).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/attachment/get/{mid}/{aid}", httphandler.DeleteAttachment).Methods(httphandlers.DELETE)
	r.HandleFunc("/smtp2/attachment/get/{mid}/{aid}", httphandler.GetAttachment).Methods(httphandlers.GET)
	r.HandleFunc("/smtp2/attachment/retrieve/{mid}/{aid}", httphandler.RetrieveAttachment).Methods(httphandlers.GET)

	// User functions
	r.HandleFunc("/user/new", httphandler.NewUser).Methods(httphandlers.POST)
	r.HandleFunc("/user/login", httphandler.Login).Methods(httphandlers.POST)

	// SMTP2 Sender Server Verification Protocol
	r.HandleFunc("/smtp2/message/verify", httphandler.MessageVerificationHandlers).Methods(httphandlers.GET)

	sugared.Info("Serving...")
	serve := config.Host + ":" + config.Port
	sugared.Info("Serving on following URL: " + serve)
	err = http.ListenAndServe(serve, r)
	if err != nil {
		sugared.Fatal(err.Error())
	}

	sugared.Info("Done serving...")
}
