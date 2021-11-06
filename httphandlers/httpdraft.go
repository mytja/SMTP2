package httphandlers

import (
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/objects"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

func (server *httpImpl) NewDraft(w http.ResponseWriter, r *http.Request) {
	ok, from, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	var originalmessage *objects.Message
	var replyid = ""
	var replypass = ""
	var originalid = -1

	replyto := r.Header.Get("ReplyTo")
	if replyto != "" {
		replytoid, err := strconv.Atoi(replyto)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusBadRequest)
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
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = server.db.CommitSentMessage(sentmsg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, fmt.Sprint(id), http.StatusCreated)
}

func (server *httpImpl) UpdateDraft(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	body := r.FormValue("Body")
	to := r.FormValue("To")
	id := r.FormValue("ID")
	idint, err := strconv.Atoi(id)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, from, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	originaldraft, err := server.db.GetSentMessage(idint)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if originaldraft.FromEmail != from {
		helpers.Write(w, "You don't have access to this resource.", http.StatusForbidden)
		return
	}

	basemsg, err := server.db.GetOriginalMessageFromOriginalID(idint)
	if err != nil {
		helpers.Write(w, "Failed to retrieve base draft message from database", http.StatusInternalServerError)
		return
	}
	if !basemsg.IsDraft {
		helpers.Write(w,
			"This message isn't a draft anymore...\nYour changes were denied, but you can use a different endpoint",
			http.StatusNotAcceptable,
		)
		return
	}

	sentmsg := sql.NewDraftSentMessage(idint, title, to, from, body)

	err = server.db.UpdateDraftSentMessage(sentmsg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "OK - Updated successfully!", http.StatusOK)
}
