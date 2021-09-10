package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mytja/SMTP2/sql"
	"net/http"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := Write(w, "Hello!\nSMTP2 is running", http.StatusOK)
	if err != nil {
		fmt.Println(err)
	}
}

func NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	q := r.Header
	title := q.Get("Title")
	uri := q.Get("URI")
	to := q.Get("To")
	from := q.Get("From")
	msg := sql.NewMessage(title, uri, to, from)
	err := sql.DB.CommitMessages(msg)
	if err != nil {
		Write(w, err.Error(), 500)
		return
	}
	Write(w, "Created", 204)
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	fmt.Println(username)
	inbox, err := sql.DB.GetInbox(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	var m []map[string]string
	for i := 0; i < len(inbox); i++ {
		var m1 = make(map[string]string)
		var msg sql.Message = inbox[i]
		m1["URI"] = msg.URI
		m1["To"] = msg.ToEmail
		m1["From"] = msg.FromEmail
		m1["Title"] = msg.Title

		m = append(m, m1)
	}
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	_, err = Write(w, string(response[:]), 200)
	if err != nil {
		return
	}
}
