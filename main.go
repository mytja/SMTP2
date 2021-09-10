package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/handlers"
	"github.com/mytja/SMTP2/sql"
	"net/http"
)

func main() {
	fmt.Println("Starting SMTP2 server...")

	if sql.DBERR != nil {
		fmt.Println("FATAL: Error while creating database")
		fmt.Println(sql.DBERR)
		return
	}

	sql.DB.Init()

	fmt.Println("INFO: Database created successfully")

	r := mux.NewRouter()
	r.HandleFunc("/smtp2", handlers.WelcomeHandler)
	r.HandleFunc("/smtp2/new/message", handlers.NewMessageHandler)
	r.HandleFunc("/smtp2/get/messages", handlers.GetMessagesHandler)

	fmt.Println("Serving...")
	err := http.ListenAndServe("127.0.0.1:8080", r)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Done serving...")
}
