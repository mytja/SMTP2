package httphandlers

import (
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

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

	replyto := r.Header.Get("ReplyTo")
	if replyto != "" {
		replytoid, err := strconv.Atoi(replyto)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to convert ReplyTo to integer", Success: false}, http.StatusBadRequest)
			return
		}

		originalmessage, err = server.db.GetOriginalMessageFromReplyTo(replytoid)
		if err != nil {
			return
		}
		replyid = originalmessage.ReplyID
		replypass = originalmessage.ReplyPass
		if originalmessage.Type == "sent" {
			originalid = originalmessage.ID
		} else {
			originalid = originalmessage.OriginalID
		}
	}

	id := server.db.GetLastMessageID()
	msg := sql.NewDraftMessage(id, originalid, replypass, replyid)
	sentmsg := sql.NewDraftSentMessage(id, "", "", from, "")
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
	WriteJSON(w, Response{Data: id, Success: true}, http.StatusCreated)
}
