package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/security"
	"github.com/mytja/SMTP2/sql"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Disable anyone with JWT to reply. (remote server does that for us - not TO DO anymore)
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
	if fromemail != server.config.HostURL {
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
	if replytomsg.Type == "sent" {
		message, err := server.db.GetSentMessage(replytoid)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve Sent message.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		to = message.ToEmail
	} else {
		message, err := server.db.GetReceivedMessage(replytoid)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve Received message.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		to = message.FromEmail
	}

	todomain, err := helpers.GetDomainFromEmail(to)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve domain from to email", Success: false}, http.StatusBadRequest)
		return
	}

	// Check if website has enabled HTTPS
	resp, err := http.Get("http://" + todomain + "/smtp2/server/info")
	if err != nil {
		WriteJSON(w, Response{Data: "Remote server isn't avaiable at the moment. Failed to send message.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
	}
	reqbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to read response body", Success: false}, http.StatusInternalServerError)
		return
	}
	var resjson ServerInfo
	err = json.Unmarshal(reqbody, &resjson)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to unmarshal remote server's response.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	protocol := "http://"
	if resjson.HasHTTPS {
		protocol = "https://"
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

	reqdom := protocol + todomain + "/smtp2/message/receive"
	req, err := http.NewRequest("POST", reqdom, strings.NewReader(""))
	req.Header.Set("Title", title)
	req.Header.Set("To", to)
	req.Header.Set("From", from)
	req.Header.Set("ServerPass", pass)
	req.Header.Set("ReplyPass", replytomsg.ReplyPass)
	req.Header.Set("ReplyID", replytomsg.ReplyID)
	req.Header.Set("ServerID", fmt.Sprint(id))
	req.Header.Set("OriginalID", fmt.Sprint(originalid))
	req.Header.Set("MVPPass", fmt.Sprint(mvppass))
	req.Header.Set(
		"URI",
		urlprotocol+server.config.HostURL+"/smtp2/message/get/"+fmt.Sprint(id)+"?pass="+pass,
	)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to request to remote server", Success: false}, http.StatusForbidden)
		return
	}
	if res.StatusCode == http.StatusCreated {
		// And let's make a 201 response
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusCreated)
		return
	}

	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		WriteJSON(w, Response{Data: "Error while reading request body", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if res.StatusCode == http.StatusNotAcceptable {
		server.logger.Error("Server has denied message")
		WriteJSON(w, Response{Data: helpers.BytearrayToString(body2), Success: false}, http.StatusNotAcceptable)
		server.db.DeleteMessage(basemsg.ID)
		server.db.DeleteSentMessage(reply.ID)
		return
	}
	server.logger.Info(req.Header.Get("URI"))
	server.logger.Info(reqdom)
	if constants.EnableDeletingOnUnknownError {
		server.db.DeleteMessage(basemsg.ID)
		server.db.DeleteSentMessage(reply.ID)
	}
	WriteJSON(w, Response{Data: res.StatusCode, Error: "Unknown error: " + helpers.BytearrayToString(body2), Success: false}, http.StatusInternalServerError)
}
