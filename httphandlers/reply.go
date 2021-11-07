package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/helpers/constants"
	"github.com/mytja/SMTP2/objects"
	"github.com/mytja/SMTP2/security"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
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
	ok, from, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	fromemail, err := helpers.GetDomainFromEmail(from)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fromemail != server.config.HostURL {
		helpers.Write(w, "This server doesn't hold your domain.", http.StatusForbidden)
		return
	}
	server.logger.Info(fmt.Sprint(fromemail, " ", server.config.HostURL))

	replytoid, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}

	replytomsg, err := server.db.GetMessageFromReplyTo(replytoid)
	if err != nil {
		helpers.Write(w, "Failed retrieving original message", http.StatusInternalServerError)
		return
	}

	// Here we get either SentMessage or ReceivedMessage
	var to string
	if replytomsg.Type == "sent" {
		message, err := server.db.GetSentMessage(replytoid)
		if err != nil {
			helpers.Write(w, "Failed to retrieve Sent message.", http.StatusInternalServerError)
			return
		}
		to = message.ToEmail
	} else {
		message, err := server.db.GetReceivedMessage(replytoid)
		if err != nil {
			helpers.Write(w, "Failed to retrieve Received message.", http.StatusInternalServerError)
			return
		}
		to = message.FromEmail
	}

	todomain, err := helpers.GetDomainFromEmail(to)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if website has enabled HTTPS
	resp, err := http.Get("http://" + todomain + "/smtp2/server/info")
	if err != nil {
		helpers.Write(w, "Remote server isn't avaiable at the moment. Failed to send message.", http.StatusInternalServerError)
	}
	reqbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resjson := make(map[string]interface{})
	err = json.Unmarshal(reqbody, &resjson)
	if err != nil {
		helpers.Write(w, "Failed to unmarshal remote server's response.", http.StatusInternalServerError)
		return
	}
	protocol := "http://"
	if resjson["hasHTTPS"] == true {
		protocol = "https://"
	}

	var originalid int
	if replytomsg.OriginalID == -1 {
		originalid = replytomsg.ID
	} else {
		originalid = replytomsg.OriginalID
	}

	id := server.db.GetLastMessageID()
	basemsg := objects.NewMessage(id, originalid, -1, replytomsg.ReplyPass, replytomsg.ReplyID, "sent", false)
	err = server.db.CommitMessage(basemsg)
	if err != nil {
		helpers.Write(w, "Failed while committing message base", http.StatusInternalServerError)
		return
	}

	// Generate random password
	pass, err := security.GenerateRandomString(80)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mvppass, err := security.GenerateRandomString(80)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reply := sql.NewSentMessage(id, title, to, from, body, pass, mvppass)
	server.logger.Info(reply)
	err = server.db.CommitSentMessage(reply)
	if err != nil {
		helpers.Write(w, "Failed to commit Sent message", http.StatusInternalServerError)
		return
	}

	// Now let's send a request to a recipient email server
	server.logger.Info(todomain)

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

	//time.Sleep(1 * time.Second)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusForbidden)
		return
	}
	if res.StatusCode == http.StatusCreated {
		// And let's make a 201 response
		helpers.Write(w, "OK", http.StatusCreated)
		return
	}

	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		helpers.Write(w, "Error while reading request body", http.StatusInternalServerError)
		return
	}
	if res.StatusCode == http.StatusNotAcceptable {
		server.logger.Error("Server has denied message")
		helpers.Write(w, helpers.BytearrayToString(body2), http.StatusNotAcceptable)
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
	helpers.Write(w, "Unknown error: "+fmt.Sprint(res.StatusCode)+" - "+helpers.BytearrayToString(body2), http.StatusInternalServerError)
}
