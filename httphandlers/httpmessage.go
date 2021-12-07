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

func (server *httpImpl) GetSentMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	pass := r.URL.Query().Get("pass")
	if pass == "" {
		WriteJSON(w, Response{Data: "Bad request - pass wasn't specified", Success: false}, http.StatusBadRequest)
		return
	}
	message, err := server.db.GetSentMessage(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Message doesn't exist or internal server error", Error: err.Error(), Success: false}, http.StatusNotFound)
		return
	}
	basemessage, err := server.db.GetOriginalMessageFromOriginalID(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve base message from database", Error: err.Error(), Success: false}, http.StatusNotFound)
		return
	}
	if basemessage.IsDraft {
		// Top secret on a FBI level - highly classified, you can go to jail if you tell anybody about this :wink: :rofl:
		WriteJSON(w, Response{Data: "This message is a draft and therefore, we should not give you any information.", Success: false}, http.StatusForbidden)
		return
	}
	if message.Pass != pass {
		WriteJSON(w, Response{Data: "Could not confirm Message password", Success: false}, http.StatusForbidden)
		return
	}

	protocol := "http://"
	if server.config.HTTPSEnabled {
		protocol = "https://"
	}

	attachments, err := server.db.GetAllAttachments(id)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve attachments from database", Success: false}, http.StatusInternalServerError)
		return
	}
	var attachmentsmap = ReceivedMessageResponse{Data: ReceivedMessageData{
		ID:          message.ID,
		Title:       message.Title,
		Receiver:    message.ToEmail,
		Sender:      message.FromEmail,
		Body:        message.Body,
		Attachments: make([]Attachment, 0),
	}}
	for i := 0; i < len(attachments); i++ {
		att := attachments[i]
		var attachment = Attachment{
			ID:       att.ID,
			Filename: att.OriginalName,
			URL:      protocol + server.config.HostURL + "/smtp2/attachment/retrieve/" + fmt.Sprint(message.ID) + "/" + fmt.Sprint(att.ID) + "?pass=" + att.AttachmentPass,
		}
		attachmentsmap.Data.Attachments = append(attachmentsmap.Data.Attachments, attachment)
	}
	WriteJSON(w, attachmentsmap, http.StatusOK)
}

func (server *httpImpl) GetInboxHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := crypto2.CheckUser(r)
	if err != nil || !isAuth {
		WriteForbiddenJWT(w, err)
		return
	}
	server.logger.Info(username)
	inbox, err := server.db.GetInbox(username)
	if err != nil {
		server.logger.Info(err)
		return
	}
	protocol := "http://"
	if server.config.HTTPSEnabled {
		protocol = "https://"
	}
	var m = InboxDataResponse{Data: make([]MessageData, 0)}
	for i := 0; i < len(inbox); i++ {
		msg := inbox[i]
		var m1 = MessageData{
			ID:       msg.ID,
			URI:      protocol + server.config.HostURL + "/smtp2/message/retrieve/" + fmt.Sprint(msg.ID),
			Receiver: msg.ToEmail,
			Sender:   msg.FromEmail,
			Title:    msg.Title,
		}
		m.Data = append(m.Data, m1)
	}
	WriteJSON(w, m, http.StatusOK)
}

func (server *httpImpl) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	body := r.FormValue("Body")
	to := r.FormValue("To")
	id := r.FormValue("ID")
	idint, err := strconv.Atoi(id)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to convert ID to integer", Success: false}, http.StatusBadRequest)
		return
	}

	ok, from, err := crypto2.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
	}

	originaldraft, err := server.db.GetSentMessage(idint)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve SentMessage from database", Success: false}, http.StatusInternalServerError)
		return
	}

	if originaldraft.FromEmail != from {
		WriteJSON(w, Response{Data: "You don't have access to this resource.", Success: false}, http.StatusForbidden)
		return
	}

	basemsg, err := server.db.GetOriginalMessageFromOriginalID(idint)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve base draft message from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if !basemsg.IsDraft {
		server.logger.Info("Ignored To, as you can't change it, because Message is already sent.")
		to = originaldraft.ToEmail
	}

	sentmsg := sql.NewDraftSentMessage(idint, title, to, from, body)

	err = server.db.UpdateDraftSentMessage(sentmsg)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false, Data: "[FATAL] Failed to update draft message"}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "OK - Updated successfully!", Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteJSON(w, Response{Data: "Not a valid integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	ok, from, err := crypto2.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	message, err := server.db.GetSentMessage(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve Sent message from database", Error: err.Error(), Success: false}, http.StatusNotFound)
		return
	}

	if message.FromEmail != from {
		WriteJSON(w, Response{Data: "You don't have permission to delete this message", Success: false}, http.StatusForbidden)
		return
	}

	server.db.DeleteSentMessage(id)
	server.db.DeleteMessage(id)

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) RetrieveMessageFromRemoteServer(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := crypto2.CheckUser(r)
	if err != nil || !isAuth {
		WriteForbiddenJWT(w, err)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteJSON(w, Response{Data: "Not a valid integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}
	msg, err := server.db.GetReceivedMessage(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve message from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if msg.ToEmail != username {
		WriteJSON(w, Response{Data: "This message wasn't intended for you!", Success: false}, http.StatusForbidden)
		return
	}
	resp, err := http.Get(msg.URI)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to make a request", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to read response body", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	bodystring := helpers.BytearrayToString(body)
	if resp.StatusCode != http.StatusOK {
		WriteJSON(w, Response{Data: bodystring, Success: false}, resp.StatusCode)
		return
	}
	server.logger.Debug(bodystring)

	// Let's manipulate string to hide URLs to attachments
	// TLDR: Some -advanced- HIGH TECH manipulation
	var j ReceivedMessageResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		server.logger.Debug(err)
		WriteJSON(w, Response{Data: "Failed to unmarshal request data", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var attachments = make([]Attachment, 0)
	var att = j.Data.Attachments
	for i := 0; i < len(att); i++ {
		attachment := att[i]
		protocol := "http://"
		if server.config.HTTPSEnabled {
			protocol = "https://"
		}
		url := protocol + server.config.HostURL + "/smtp2/attachment/remote/get/" + fmt.Sprint(id) + "/" + fmt.Sprint(attachment.ID)
		attachment.URL = url
		attachments = append(attachments, attachment)
	}
	j.Data.Attachments = attachments
	WriteJSON(w, j, http.StatusOK)
}
