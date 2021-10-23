package httphandlers

import (
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/objects"
	"github.com/mytja/SMTP2/security"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

func ReceiveMessageHandler(w http.ResponseWriter, r *http.Request) {
	q := r.Header
	title := q.Get("Title")
	uri := q.Get("URI")
	to := q.Get("To")
	from := q.Get("From")
	id := q.Get("ServerID")
	pass := q.Get("ServerPass")
	replyid := q.Get("ReplyPass")
	replypass := q.Get("ReplyID")
	if replyid == "" || replypass == "" {
		helpers.Write(w, "Bad request", http.StatusBadRequest)
		return
	}
	fmt.Println(id, title, uri, to, from)
	atoi, err := strconv.Atoi(id)
	if err != nil {
		helpers.Write(w, "ID isn't a valid integer", http.StatusBadRequest)
		return
	}
	_, err = sql.DB.GetUserByEmail(to)
	if err != nil {
		helpers.Write(w,
			fmt.Sprint("Could not find recipient in our database or there was an internal error in recipient's database: ", err.Error()),
			http.StatusNotAcceptable,
		)
		return
	}
	msgid := sql.DB.GetLastMessageID()
	// TODO: DO NOT MANUALLY INSERT -1
	basemsg := objects.NewMessage(msgid, -1, atoi, replypass, replyid, "received")

	msg := sql.NewReceivedMessage(title, uri, to, from, atoi, pass)
	msg.ID = msgid

	verification, _ := security.VerifyMessage(msg)
	if !verification {
		helpers.Write(w, "Failed to verify message.", http.StatusForbidden)
		return
	}
	err = sql.DB.CommitMessage(basemsg)
	if err != nil {
		helpers.Write(w, fmt.Sprint("Failed commiting base message to database", err.Error()), http.StatusInternalServerError)
		return
	}
	err = sql.DB.CommitReceivedMessages(msg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "Created", http.StatusCreated)
}
