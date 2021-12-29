package httphandlers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/imroc/req"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/security"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

func (server *httpImpl) NewReplyHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	body := r.FormValue("Body")
	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	fromemail, err := helpers.GetDomainFromEmail(from)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve domain from email", Success: false}, http.StatusInternalServerError)
		return
	}
	if fromemail != server.config.HostURL && !server.config.SkipSameDomainCheck {
		WriteJSON(w, Response{Data: "This server doesn't hold your domain.", Success: false}, http.StatusForbidden)
		return
	}
	server.logger.Debug(fmt.Sprint(fromemail, " ", server.config.HostURL))

	replytoid, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "ID isn't a integer", Success: false}, http.StatusBadRequest)
		return
	}

	replytomsg, err := server.db.GetMessageFromReplyTo(replytoid)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed retrieving original message", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	// Here we get either SentMessage or ReceivedMessage
	var to string
	var from_email string
	if replytomsg.Type == "sent" {
		message, err := server.db.GetSentMessage(replytoid)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve Sent message.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		to = message.ToEmail
		from_email = message.FromEmail
	} else {
		message, err := server.db.GetReceivedMessage(replytoid)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve Received message.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		to = message.FromEmail
		from_email = message.ToEmail
	}

	if from_email != from {
		WriteForbiddenJWT(w, errors.New("you don't own this message that you are replying to"))
		return
	}

	todomain, err := helpers.GetDomainFromEmail(to)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve domain from to email", Success: false}, http.StatusBadRequest)
		return
	}

	// Check if website has enabled HTTPS
	protocol, err := server.security.GetProtocolFromDomain(todomain)
	if err != nil {
		WriteJSON(w, Response{Data: "Remote server isn't avaiable at the moment. Failed to send message.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
	}

	var originalid int
	if replytomsg.OriginalID == -1 {
		originalid = replytomsg.ID
	} else {
		originalid = replytomsg.OriginalID
	}

	id := server.db.GetLastMessageID()
	basemsg := sql.NewMessage(id, originalid, -1, replytomsg.ReplyPass, replytomsg.ReplyID, "sent", false)
	err = server.db.CommitMessage(basemsg)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while committing message base", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	// Generate random password
	pass, err := security.GenerateRandomString(80)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to generate random string", Success: false}, http.StatusInternalServerError)
		return
	}
	mvppass, err := security.GenerateRandomString(80)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to generate random string", Success: false}, http.StatusInternalServerError)
		return
	}

	reply := sql.NewSentMessage(id, title, to, from, body, pass, mvppass)
	server.logger.Debug(reply)
	err = server.db.CommitSentMessage(reply)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to commit Sent message", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	// Now let's send a request to a recipient email server
	server.logger.Debug(todomain)

	urlprotocol := "http://"
	if server.config.HTTPSEnabled {
		urlprotocol = "https://"
	}

	mailurl := urlprotocol + server.config.HostURL + "/smtp2/message/get/" + fmt.Sprint(id) + "?pass=" + pass

	headers := SentMessage{
		Title:      reply.Title,
		To:         reply.ToEmail,
		From:       reply.FromEmail,
		ServerPass: reply.Pass,
		ReplyPass:  basemsg.ReplyPass,
		ReplyID:    basemsg.ReplyID,
		OriginalID: fmt.Sprint(originalid),
		ServerID:   fmt.Sprint(id),
		MVPPass:    mvppass,
		URI:        mailurl,
	}

	reqdom := protocol + todomain + "/smtp2/message/receive"
	res, err := req.Post(reqdom, req.HeaderFromStruct(headers))
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to request to remote server", Success: false}, http.StatusForbidden)
		return
	}

	code := res.Response().StatusCode
	resbody := res.String()

	if code == http.StatusCreated {
		// And let's make a 201 response
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
		return
	}

	if code == http.StatusNotAcceptable {
		server.logger.Error("Server has denied message")
		WriteJSON(w, Response{Data: resbody, Success: false}, http.StatusNotAcceptable)
		server.db.DeleteMessage(basemsg.ID)
		server.db.DeleteSentMessage(reply.ID)
		return
	}
	server.logger.Debugw("message details", "mail_url", mailurl, "domain", reqdom)
	if constants.EnableDeletingOnUnknownError {
		server.db.DeleteMessage(basemsg.ID)
		server.db.DeleteSentMessage(reply.ID)
	}
	WriteJSON(w, Response{Data: code, Error: fmt.Sprint("Unknown error: ", resbody), Success: false}, http.StatusInternalServerError)
}
