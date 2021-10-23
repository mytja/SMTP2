package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

func GetReceivedMessageHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, email, err := crypto2.CheckUser(r)
	if isAuth == false {
		helpers.Write(w, "unauthenticated", http.StatusForbidden)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}
	message, err := sql.DB.GetReceivedMessage(id)
	if err != nil {
		helpers.Write(w, "Message doesn't exist or internal server error: "+err.Error(), http.StatusNotFound)
		return
	}
	if message.ToEmail != email {
		helpers.Write(w, "unauthenticated", http.StatusForbidden)
	}
	var m map[string]string
	m["ID"] = fmt.Sprint(message.ID)
	m["ServerID"] = fmt.Sprint(message.ServerID)
	m["Title"] = message.Title
	m["URI"] = message.URI
	m["ServerPass"] = message.ServerPass
	m["Receiver"] = message.ToEmail
	m["Sender"] = message.FromEmail
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
}

func GetSentMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	pass := r.URL.Query().Get("pass")
	if pass == "" {
		helpers.Write(w, "Bad request - pass wasn't specified", http.StatusBadRequest)
		return
	}
	message, err := sql.DB.GetSentMessage(id)
	if message.Pass != pass {
		helpers.Write(w, "Could not confirm Message password", http.StatusForbidden)
		return
	}
	if err != nil {
		helpers.Write(w, "Message doesn't exist or internal server error: "+err.Error(), http.StatusNotFound)
		return
	}
	var m = make(map[string]string)
	m["ID"] = fmt.Sprint(message.ID)
	m["Title"] = message.Title
	m["Pass"] = message.Pass
	m["Receiver"] = message.ToEmail
	m["Sender"] = message.FromEmail
	m["Body"] = message.Body
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
}

func GetInboxHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	} else if isAuth == false {
		helpers.Write(w, "Not authenticated", http.StatusForbidden)
		return
	}
	fmt.Println(username)
	inbox, err := sql.DB.GetInbox(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	var m []map[string]string
	for i := 0; i < len(inbox); i++ {
		var m1 = make(map[string]string)
		var msg sql.ReceivedMessage = inbox[i]
		m1["URI"] = msg.URI
		m1["To"] = msg.ToEmail
		m1["From"] = msg.FromEmail
		m1["Title"] = msg.Title

		m = append(m, m1)
	}
	if m == nil {
		m = make([]map[string]string, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	_, err = helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
	if err != nil {
		return
	}
}
