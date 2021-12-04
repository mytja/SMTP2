package httphandlers

import (
	"fmt"
	"github.com/jpillora/go-tld"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/objects"
	"github.com/mytja/SMTP2/security"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

func (server *httpImpl) ReceiveMessageHandler(w http.ResponseWriter, r *http.Request) {
	q := r.Header
	title := q.Get("Title")
	uri := q.Get("URI")
	to := q.Get("To")
	from := q.Get("From")
	id := q.Get("ServerID")
	pass := q.Get("ServerPass")
	replyid := q.Get("ReplyPass")
	replypass := q.Get("ReplyID")
	originalid := q.Get("OriginalID")
	mvppass := q.Get("MVPPass")
	if replyid == "" || replypass == "" || mvppass == "" {
		WriteJSON(w, Response{Data: "Bad request", Success: false}, http.StatusBadRequest)
		return
	}
	server.logger.Info(fmt.Sprint(id, title, uri, to, from, mvppass))
	atoi, err := strconv.Atoi(id)
	if err != nil {
		WriteJSON(w, Response{Data: "ID isn't a valid integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}
	originalidint, err := strconv.Atoi(originalid)
	if err != nil {
		WriteJSON(w, Response{Data: "OriginalID isn't a valid integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	var isOriginal = false

	originalmessage, err := server.db.GetOriginalFromReplyHeaders(replyid, replypass)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			originalidint = -1
			isOriginal = true
		} else {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve original message from your reply headers", Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if !isOriginal {
		if originalmessage.Type == "sent" {
			originalmsg, err := server.db.GetSentMessage(originalmessage.ID)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Unable to retrieve original message from SentMessages.", Success: false}, http.StatusInternalServerError)
				return
			}
			if !(originalmsg.ToEmail == from && originalmsg.FromEmail == to) {
				WriteJSON(w, Response{Data: "You didn't send this message...", Success: false}, http.StatusNotAcceptable)
				return
			}
		} else {
			originalmsg, err := server.db.GetReceivedMessage(originalmessage.ID)
			if err != nil {
				server.logger.Info(err)
				WriteJSON(w, Response{Data: "Unable to retrieve original message from ReceivedMessages.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			if !(originalmsg.ToEmail == to && originalmsg.FromEmail == from) {
				WriteJSON(w, Response{Data: "You didn't send this message...", Success: false}, http.StatusNotAcceptable)
				return
			}
		}
	}
	_, err = server.db.GetUserByEmail(to)
	if err != nil {
		WriteJSON(w,
			Response{Data: "Could not find recipient in our database or there was an internal error in recipient's database: ", Error: err.Error(), Success: false},
			http.StatusNotAcceptable,
		)
		return
	}
	msgid := server.db.GetLastMessageID()
	basemsg := objects.NewMessage(msgid, originalidint, atoi, replypass, replyid, "received", false)

	msg := sql.NewReceivedMessage(msgid, title, uri, to, from, atoi, pass, "", mvppass)

	verification, err := security.VerifyMessage(msg)
	if !verification {
		server.logger.Infow("failed to verify", "error", err)
		WriteJSON(w, Response{Data: "Failed to verify message.", Success: false}, http.StatusForbidden)
		return
	}
	domain, err := helpers.GetDomainFromEmail(from)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve domain from FromEmail", Success: false}, http.StatusBadRequest)
		return
	}
	fromurl, err := tld.Parse(domain)
	if err != nil {
		msg.Warning = "SMTP2_FAILED_TO_PARSE_EMAIL_ADDRESS_AS_URI"
	}
	requrl, err := tld.Parse(uri)
	if err != nil {
		msg.Warning = "SMTP2_FAILED_TO_PARSE_DOMAIN_AS_URI"
	}
	if requrl != nil && fromurl != nil {
		server.logger.Info(fromurl.Domain)
		if requrl.Domain != fromurl.Domain {
			msg.Warning = "SMTP2_DOMAINS_NOT_MATCHING"
		}
	}

	// Commit
	err = server.db.CommitMessage(basemsg)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed commiting base message to database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	err = server.db.CommitReceivedMessages(msg)
	if err != nil {
		server.logger.Error("Failed to commit received message")
		WriteJSON(w, Response{Data: "[FATAL] Failed to commit received message", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "Created", http.StatusCreated)
}
