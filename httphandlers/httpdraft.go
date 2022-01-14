package httphandlers

import (
	"fmt"
	"github.com/imroc/req"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
	"strings"
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
	var forwardmessage = ""

	user, err := server.db.GetUserByEmail(from)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve User from database", Success: false}, http.StatusNotFound)
		return
	}

	replyto := r.Header.Get("ReplyTo")
	usemessage := r.Header.Get("UseMessage")
	if replyto != "" && usemessage == "" {
		// Reply to message
		replytoid, err := strconv.Atoi(replyto)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to convert ReplyTo to integer", Success: false}, http.StatusBadRequest)
			return
		}

		originalmessage, err = server.db.GetMessageFromReplyTo(replytoid)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Could not find base message", Success: false}, http.StatusNotFound)
			return
		}

		replyid = originalmessage.ReplyID
		replypass = originalmessage.ReplyPass
		if originalmessage.Type == "sent" {
			originalid = originalmessage.ID
			msg, err := server.db.GetSentMessage(originalid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Could not find sent message", Success: false}, http.StatusNotFound)
				return
			}
			if msg.FromEmail != from {
				WriteForbiddenJWT(w, nil)
				return
			}
			to = msg.ToEmail
			title = fmt.Sprint("RE: ", msg.Title)
		} else {
			originalid = originalmessage.OriginalID
			msg, err := server.db.GetReceivedMessage(originalmessage.ID)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Could not find received message", Success: false}, http.StatusNotFound)
				return
			}
			if msg.ToEmail != from {
				WriteForbiddenJWT(w, nil)
				return
			}
			to = msg.FromEmail
			title = fmt.Sprint("RE: ", msg.Title)
		}
	} else if usemessage != "" && replyto == "" {
		// Forward message
		msgid, err := strconv.Atoi(usemessage)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to convert ReplyTo to integer", Success: false}, http.StatusBadRequest)
			return
		}

		originalmessage, err = server.db.GetMessageFromReplyTo(msgid)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Could not find base message", Success: false}, http.StatusNotFound)
			return
		}

		if originalmessage.Type == "sent" {
			msg, err := server.db.GetSentMessage(msgid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Could not find sent message", Success: false}, http.StatusNotFound)
				return
			}
			if msg.FromEmail != from {
				WriteForbiddenJWT(w, nil)
				return
			}
			msgbody := strings.Split(msg.Body, "\n")
			editedmsgbody := make([]string, 0)
			for i := 0; i < len(msgbody); i++ {
				line := msgbody[i]
				editedmsgbody = append(editedmsgbody, fmt.Sprint("> ", line))
			}

			// This formatting is disgusting, but it's currently best possible

			forwardmessage = fmt.Sprintf(`
> # Forwarded message
> Warning: THIS MESSAGE IS A SNAPSHOT OF STATE WHEN MESSAGE WAS SENT - IT DOESN'T NECESSARILY REPRESENT CURRENT STATE OF THE MESSAGE.
> 
> To: %s
> From: %s
> Subject: %s
> Body:
%s
			`, msg.ToEmail, msg.FromEmail, msg.Title, strings.Join(editedmsgbody, "\n"))
		} else {
			msg, err := server.db.GetReceivedMessage(msgid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Could not find received message", Success: false}, http.StatusNotFound)
				return
			}
			if msg.ToEmail != from {
				WriteForbiddenJWT(w, nil)
				return
			}

			resp, err := req.Get(msg.URI)
			if err != nil {
				WriteJSON(w, Response{Data: "Failed to make a request", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			code := resp.Response().StatusCode
			var j ReceivedMessageResponse
			err = resp.ToJSON(&j)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Data: "Failed to unmarshal response", Success: false}, http.StatusInternalServerError)
			}

			if code != http.StatusOK {
				WriteJSON(w, Response{Data: j, Success: false, Error: "Request failed on server side"}, code)
				return
			}
			server.logger.Debug(j)

			msgbody := strings.Split(j.Data.Body, "\n")
			editedmsgbody := make([]string, 0)
			for i := 0; i < len(msgbody); i++ {
				line := msgbody[i]
				editedmsgbody = append(editedmsgbody, fmt.Sprint("> ", line))
			}

			forwardmessage = fmt.Sprintf(`
> # Forwarded message
> Warning: THIS MESSAGE IS A SNAPSHOT OF STATE WHEN MESSAGE WAS SENT - IT DOESN'T NECESSARILY REPRESENT CURRENT STATE OF THE MESSAGE.
> 
> To: %s
> From: %s
> Subject: %s
> Body:
%s
			`, msg.ToEmail, msg.FromEmail, msg.Title, strings.Join(editedmsgbody, "\n"))
		}
	}

	id := server.db.GetLastMessageID()
	msg := sql.NewDraftMessage(id, originalid, replypass, replyid)
	sentmsg := sql.NewDraftSentMessage(id, title, to, from, fmt.Sprint("", "\n", user.Signature, "\n", forwardmessage))
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
