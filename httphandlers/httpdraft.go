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

func NewDraft(w http.ResponseWriter, r *http.Request) {
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

		originalmessage, err = sql.DB.GetOriginalMessageFromReplyTo(replytoid)
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

	id := sql.DB.GetLastMessageID()
	msg := sql.NewDraftMessage(id, originalid, replypass, replyid)
	sentmsg := sql.NewDraftSentMessage(id, "", "", from, "")
	err = sql.DB.CommitMessage(msg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = sql.DB.CommitSentMessage(sentmsg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, fmt.Sprint(id), http.StatusCreated)
}

func UpdateDraft(w http.ResponseWriter, r *http.Request) {
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

	originaldraft, err := sql.DB.GetSentMessage(idint)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if originaldraft.FromEmail != from {
		helpers.Write(w, "You don't have access to this resource.", http.StatusForbidden)
		return
	}

	sentmsg := sql.NewDraftSentMessage(idint, title, to, from, body)

	err = sql.DB.UpdateDraftSentMessage(sentmsg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "OK - Updated successfully!", http.StatusOK)
}
