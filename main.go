package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/httphandlers"
	"github.com/mytja/SMTP2/sql"
	"github.com/spf13/cobra"
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
	command.Flags().StringVar(&constants.SERVER_URL, "url", "http://0.0.0.0:8080", "set server URL")
	command.Flags().StringVar(&constants.DB_NAME, "dbname", "smtp2.db", "set DB name")

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sql.DB, sql.DBERR = sql.NewSQL()
	sql.DB.Init()

	if sql.DBERR != nil {
		fmt.Println("FATAL: Error while creating database")
		fmt.Println(sql.DBERR)
		return
	}

	fmt.Println("INFO: Database created successfully")

	r := mux.NewRouter()
	r.HandleFunc("/smtp2", httphandlers.WelcomeHandler).Methods(httphandlers.GET)

	// Message
	r.HandleFunc("/smtp2/message/receive", httphandlers.ReceiveMessageHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/new", httphandlers.NewMessageHandler).Methods(httphandlers.POST)
	r.HandleFunc("/smtp2/message/get", httphandlers.GetMessagesHandler).Methods(httphandlers.GET)

	// User functions
	r.HandleFunc("/user/new", httphandlers.NewUser).Methods(httphandlers.POST)
	r.HandleFunc("/user/login", httphandlers.Login).Methods(httphandlers.POST)

	// SMTP2 Sender Server Verification Protocol
	r.HandleFunc("/smtp2/message/verify", httphandlers.MessageVerificationHandlers).Methods(httphandlers.GET)

	fmt.Println("Serving...")
	serve := config.Host + ":" + config.Port
	fmt.Println(serve)
	err := http.ListenAndServe(serve, r)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Done serving...")
}
