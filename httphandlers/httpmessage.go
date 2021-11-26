package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (server *httpImpl) GetReceivedMessageHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, email, err := crypto2.CheckUser(r)
	if isAuth == false {
		helpers.Write(w, "unauthenticated", http.StatusForbidden)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}
	message, err := server.db.GetReceivedMessage(id)
	if err != nil {
		helpers.Write(w, "Message doesn't exist or internal server error: "+err.Error(), http.StatusNotFound)
		return
	}
	if message.ToEmail != email {
		helpers.Write(w, "unauthenticated", http.StatusForbidden)
		return
	}
	protocol := "http://"
	if server.config.HTTPSEnabled {
		protocol = "https://"
	}
	var m = make(map[string]string)
	m["ID"] = fmt.Sprint(message.ID)
	m["ServerID"] = fmt.Sprint(message.ServerID)
	m["Title"] = message.Title
	m["URI"] = protocol + server.config.HostURL + "/smtp2/message/retrieve/" + fmt.Sprint(message.ID)
	m["ServerPass"] = message.ServerPass
	m["Receiver"] = message.ToEmail
	m["Sender"] = message.FromEmail
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
}

func (server *httpImpl) GetSentMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	pass := r.URL.Query().Get("pass")
	if pass == "" {
		helpers.Write(w, "Bad request - pass wasn't specified", http.StatusBadRequest)
		return
	}
	message, err := server.db.GetSentMessage(id)
	if err != nil {
		helpers.Write(w, "Message doesn't exist or internal server error: "+err.Error(), http.StatusNotFound)
		return
	}
	basemessage, err := server.db.GetOriginalMessageFromOriginalID(id)
	if err != nil {
		helpers.Write(w, "Failed to retrieve base message from database: "+err.Error(), http.StatusNotFound)
		return
	}
	if basemessage.IsDraft {
		helpers.Write(w, "This message is a draft and therefore, we should not give you any information.", http.StatusForbidden)
		return
	}
	if message.Pass != pass {
		helpers.Write(w, "Could not confirm Message password", http.StatusForbidden)
		return
	}

	protocol := "http://"
	if server.config.HTTPSEnabled {
		protocol = "https://"
	}

	attachments, err := server.db.GetAllAttachments(id)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var attachmentsmap = make([]map[string]interface{}, 0)
	for i := 0; i < len(attachments); i++ {
		att := attachments[i]
		var attachment = make(map[string]interface{})
		attachment["ID"] = att.ID
		attachment["Filename"] = att.ID
		attachment["URL"] = protocol + server.config.HostURL + "/smtp2/attachment/retrieve/" + fmt.Sprint(message.ID) + "/" + fmt.Sprint(att.ID) + "?pass=" + att.AttachmentPass
		attachmentsmap = append(attachmentsmap, attachment)
	}
	var m = make(map[string]interface{})
	m["ID"] = fmt.Sprint(message.ID)
	m["Title"] = message.Title
	m["Receiver"] = message.ToEmail
	m["Sender"] = message.FromEmail
	m["Body"] = message.Body
	m["Attachments"] = attachmentsmap
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
}

func (server *httpImpl) GetInboxHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	} else if isAuth == false {
		helpers.Write(w, "Not authenticated", http.StatusForbidden)
		return
	}
	server.logger.Info(username)
	inbox, err := server.db.GetInbox(username)
	if err != nil {
		server.logger.Info(err)
		return
	}
	var m []map[string]string
	for i := 0; i < len(inbox); i++ {
		var m1 = make(map[string]string)
		var msg sql.ReceivedMessage = inbox[i]
		m1["URI"] = msg.URI
		m1["To"] = msg.ToEmail
		m1["From"] = msg.FromEmail
		m1["Title"] = msg.Title
		m1["ID"] = fmt.Sprint(msg.ID)

		m = append(m, m1)
	}
	if m == nil {
		m = make([]map[string]string, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	_, err = helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
	if err != nil {
		return
	}
}

func (server *httpImpl) UpdateMessage(w http.ResponseWriter, r *http.Request) {
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
		server.logger.Info("Ignored To, as you can't change it, because Message is already sent.")
		to = originaldraft.ToEmail
	}

	sentmsg := sql.NewDraftSentMessage(idint, title, to, from, body)

	err = server.db.UpdateDraftSentMessage(sentmsg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "OK - Updated successfully!", http.StatusOK)
}

func (server *httpImpl) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		helpers.Write(w, "Not a valid integer", http.StatusBadRequest)
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

	message, err := server.db.GetSentMessage(id)
	if err != nil {
		helpers.Write(w, "Failed to retrieve Sent message from database", http.StatusNotFound)
		return
	}

	if message.FromEmail != from {
		helpers.Write(w, "You don't have permission to delete this message", http.StatusForbidden)
		return
	}

	server.db.DeleteSentMessage(id)
	server.db.DeleteMessage(id)

	helpers.Write(w, "OK", http.StatusOK)
}

func (server *httpImpl) RetrieveMessageFromRemoteServer(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	} else if isAuth == false {
		helpers.Write(w, "Not authenticated", http.StatusForbidden)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		helpers.Write(w, "Not a valid integer", http.StatusBadRequest)
		return
	}
	msg, err := server.db.GetReceivedMessage(id)
	if err != nil {
		helpers.Write(w, "Failed to retrieve message from database", http.StatusInternalServerError)
		return
	}
	if msg.ToEmail != username {
		helpers.Write(w, "This message wasn't intended for you!", http.StatusForbidden)
		return
	}
	resp, err := http.Get(msg.URI)
	if err != nil {
		helpers.Write(w, "Failed to make a request", http.StatusInternalServerError)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		helpers.Write(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}
	bodystring := helpers.BytearrayToString(body)
	if resp.StatusCode != http.StatusOK {
		helpers.Write(w, bodystring, resp.StatusCode)
		return
	}
	helpers.Write(w, bodystring, http.StatusOK)

}
