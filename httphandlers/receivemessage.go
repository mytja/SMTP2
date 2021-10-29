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
	originalid := q.Get("OriginalID")
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
	originalidint, err := strconv.Atoi(originalid)
	if err != nil {
		helpers.Write(w, "OriginalID isn't a valid integer", http.StatusBadRequest)
		return
	}

	var isOriginal = false

	originalmessage, err := sql.DB.GetOriginalFromReplyHeaders(replyid, replypass)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			originalidint = -1
			isOriginal = true
		} else {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if !isOriginal {
		if originalmessage.Type == "sent" {
			originalmsg, err := sql.DB.GetSentMessage(originalmessage.ID)
			if err != nil {
				helpers.Write(w, "Unable to retrieve original message from SentMessages.", http.StatusInternalServerError)
				return
			}
			if !(originalmsg.ToEmail == from && originalmsg.FromEmail == to) {
				helpers.Write(w, "You didn't send this message...", http.StatusNotAcceptable)
				return
			}
		} else {
			originalmsg, err := sql.DB.GetReceivedMessage(originalmessage.ID)
			if err != nil {
				fmt.Println(err)
				helpers.Write(w, "Unable to retrieve original message from ReceivedMessages.", http.StatusInternalServerError)
				return
			}
			if !(originalmsg.ToEmail == to && originalmsg.FromEmail == from) {
				helpers.Write(w, "You didn't send this message...", http.StatusNotAcceptable)
				return
			}
		}
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
	basemsg := objects.NewMessage(msgid, originalidint, atoi, replypass, replyid, "received")

	msg := sql.NewReceivedMessage(title, uri, to, from, atoi, pass, "")
	msg.ID = msgid

	verification, _ := security.VerifyMessage(msg)
	if !verification {
		helpers.Write(w, "Failed to verify message.", http.StatusForbidden)
		return
	}
	domain, err := helpers.GetDomainFromEmail(from)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
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
		fmt.Println(fromurl.Domain)
		if requrl.Domain != fromurl.Domain {
			msg.Warning = "SMTP2_DOMAINS_NOT_MATCHING"
		}
	}

	// Commit
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
