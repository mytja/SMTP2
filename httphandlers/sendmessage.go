package httphandlers

import (
	"fmt"
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

func (server *httpImpl) NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	to := r.FormValue("To")
	body := r.FormValue("Body")
	usedraft := r.FormValue("DraftID")

	ok, from, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	var iscreatedfromdraft = false
	var originalid = -1
	var serverid = -1
	id := server.db.GetLastMessageID()

	if usedraft != "" {
		draftid, err := strconv.Atoi(usedraft)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusBadRequest)
			return
		}
		message, err := server.db.GetSentMessage(draftid)
		if err != nil {
			helpers.Write(w, "Failed to retrieve draft from database", http.StatusInternalServerError)
			return
		}
		if message.FromEmail != from {
			helpers.Write(w, "You didn't create this draft...", http.StatusForbidden)
			return
		}
		basemsg, err := server.db.GetMessageFromReplyTo(draftid)
		if err != nil {
			helpers.Write(w, "Failed to retrieve draft base from database", http.StatusInternalServerError)
			return
		}
		if !basemsg.IsDraft {
			helpers.Write(w, "This isn't a draft anymore...", http.StatusBadRequest)
			return
		}
		iscreatedfromdraft = true
		serverid = basemsg.ServerID
		originalid = basemsg.OriginalID

		title = message.Title
		body = message.Body
		to = message.ToEmail
		id = message.ID
	}

	if !strings.Contains(to, "@") {
		helpers.Write(w, "Invalid To address", http.StatusBadRequest)
		return
	}

	pass, err := security.GenerateRandomString(50)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	replyPass, err := security.GenerateRandomString(50)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	replyID, err := security.GenerateRandomString(50)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	basemsg := objects.NewMessage(id, originalid, serverid, replyPass, replyID, "sent", false)
	if iscreatedfromdraft {
		// Update instead of pushing
		err := server.db.UpdateDraftMessage(basemsg)
		if err != nil {
			helpers.Write(w, fmt.Sprint("Error while committing to database: ", err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		err = server.db.CommitMessage(basemsg)
		if err != nil {
			helpers.Write(w, fmt.Sprint("Error while committing to database: ", err.Error()), http.StatusInternalServerError)
			return
		}
	}
	server.logger.Info(basemsg)

	msg := sql.NewSentMessage(title, to, from, body, pass)
	msg.ID = id
	server.logger.Info(msg.ID)

	server.logger.Info(msg)

	// Now let's send a request to a recipient email server
	todomain, err := helpers.GetDomainFromEmail(msg.ToEmail)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}
	server.logger.Info(todomain)

	idstring := fmt.Sprint(id)
	server.logger.Info("ID2: ", idstring)

	protocol := "http://"
	if constants.ForceHttps {
		protocol = "https://"
	}
	reqdom := protocol + todomain + "/smtp2/message/receive"

	urlprotocol := "http://"
	if server.config.HTTPSEnabled {
		urlprotocol = "https://"
	}

	req, err := http.NewRequest("POST", reqdom, strings.NewReader(""))
	req.Header.Set("Title", msg.Title)
	req.Header.Set("To", msg.ToEmail)
	req.Header.Set("From", msg.FromEmail)
	req.Header.Set("ServerPass", msg.Pass)
	req.Header.Set("ReplyPass", basemsg.ReplyPass)
	req.Header.Set("ReplyID", basemsg.ReplyID)
	req.Header.Set("OriginalID", "-1")
	req.Header.Set("ServerID", fmt.Sprint(idstring))
	req.Header.Set(
		"URI",
		urlprotocol+server.config.HostURL+"/smtp2/message/get/"+fmt.Sprint(id)+"?pass="+msg.Pass,
	)

	// We have to commit a message before we send a request
	if iscreatedfromdraft {
		err = server.db.UpdateDraftSentMessage(msg)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err = server.db.CommitSentMessage(msg)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//time.Sleep(1 * time.Second)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusForbidden)
		return
	}
	body3, _ := ioutil.ReadAll(res.Body)
	server.logger.Info(helpers.BytearrayToString(body3))
	if res.StatusCode == http.StatusCreated {
		// And let's make a 201 response
		helpers.Write(w, "OK", http.StatusCreated)
		return
	}
	if res.StatusCode == http.StatusNotAcceptable {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			helpers.Write(w, "Error while reading request body", http.StatusInternalServerError)
			return
		}
		helpers.Write(w, helpers.BytearrayToString(body)+"\nMessage has been automatically deleted", http.StatusNotAcceptable)
		server.db.DeleteMessage(basemsg.ID)
		server.db.DeleteSentMessage(msg.ID)
		return
	}
	server.logger.Info(req.Header.Get("URI"))
	server.logger.Info(reqdom)
	if constants.EnableDeletingOnUnknownError {
		server.db.DeleteMessage(basemsg.ID)
		server.db.DeleteSentMessage(msg.ID)
	}
	helpers.Write(w, "Unknown error: "+fmt.Sprint(res.StatusCode), http.StatusInternalServerError)
}
