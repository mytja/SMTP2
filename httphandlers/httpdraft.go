package httphandlers

import (
	"fmt"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

type DraftResponse struct {
	ID    int    `json:"ID"`
	To    string `json:"To"`
	Title string `json:"Title"`
	Body  string `json:"Body"`
	From  string `json:"From"`
}

func (server *httpImpl) NewDraft(w http.ResponseWriter, r *http.Request) {
	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	var originalmessage *sql.Message
	var replyid = ""
	var replypass = ""
	var originalid = -1
	var to = ""
	var title = ""

	user, err := server.db.GetUserByEmail(from)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve User from database", Success: false}, http.StatusNotFound)
		return
	}

	replyto := r.Header.Get("ReplyTo")
	if replyto != "" {
		replytoid, err := strconv.Atoi(replyto)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to convert ReplyTo to integer", Success: false}, http.StatusBadRequest)
			return
		}

		originalmessage, err = server.db.GetMessageFromReplyTo(replytoid)
		if err != nil {
			return
		}

		replyid = originalmessage.ReplyID
		replypass = originalmessage.ReplyPass
		if originalmessage.Type == "sent" {
			originalid = originalmessage.ID
			msg, err := server.db.GetSentMessage(originalid)
			if err != nil {
				return
			}
			to = msg.ToEmail
			title = fmt.Sprint("RE: ", msg.Title)
		} else {
			originalid = originalmessage.OriginalID
			msg, err := server.db.GetReceivedMessage(originalmessage.ID)
			if err != nil {
				return
			}
			to = msg.FromEmail
			title = fmt.Sprint("RE: ", msg.Title)
		}
	}

	id := server.db.GetLastMessageID()
	msg := sql.NewDraftMessage(id, originalid, replypass, replyid)
	sentmsg := sql.NewDraftSentMessage(id, title, to, from, fmt.Sprint("", "\n", user.Signature))
	err = server.db.CommitMessage(msg)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit base Message to database", Success: false}, http.StatusInternalServerError)
		return
	}
	err = server.db.CommitSentMessage(sentmsg)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit SentMessage to database", Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: DraftResponse{ID: id, Title: title, To: to, Body: sentmsg.Body}, Success: true}, http.StatusCreated)
}
