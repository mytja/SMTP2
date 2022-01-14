package httphandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/imroc/req"
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
	basemessage, err := server.db.GetMessageFromReplyTo(id)
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
	isAuth, username, err := server.security.CheckUser(r)
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
			IsRead:   msg.IsRead,
		}
		m.Data = append(m.Data, m1)
	}
	WriteJSON(w, m, http.StatusOK)
}

func (server *httpImpl) GetSentInboxHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := server.security.CheckUser(r)
	if err != nil || !isAuth {
		WriteForbiddenJWT(w, err)
		return
	}
	server.logger.Info(username)

	inbox, err := server.db.GetAllSentMessages(username)
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
		basemsg, err := server.db.GetMessageFromReplyTo(msg.ID)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve base message from database", Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		if !basemsg.IsDraft {
			var m1 = MessageData{
				ID:       msg.ID,
				URI:      protocol + server.config.HostURL + "/smtp2/message/sent/get/" + fmt.Sprint(msg.ID),
				Receiver: msg.ToEmail,
				Sender:   msg.FromEmail,
				Title:    msg.Title,
			}
			m.Data = append(m.Data, m1)
		}
	}
	WriteJSON(w, m, http.StatusOK)
}

func (server *httpImpl) GetDraftInboxHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := server.security.CheckUser(r)
	if err != nil || !isAuth {
		WriteForbiddenJWT(w, err)
		return
	}
	server.logger.Info(username)

	inbox, err := server.db.GetAllSentMessages(username)
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
		basemsg, err := server.db.GetMessageFromReplyTo(msg.ID)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve base message from database", Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		if basemsg.IsDraft {
			var m1 = MessageData{
				ID:       msg.ID,
				URI:      protocol + server.config.HostURL + "/smtp2/message/sent/get/" + fmt.Sprint(msg.ID),
				Receiver: msg.ToEmail,
				Sender:   msg.FromEmail,
				Title:    msg.Title,
			}
			m.Data = append(m.Data, m1)
		}
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

	ok, from, err := server.security.CheckUser(r)
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

	basemsg, err := server.db.GetMessageFromReplyTo(idint)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve base draft message from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if !basemsg.IsDraft {
		server.logger.Info("Ignored To, as you can't change it, because Message is already sent.")
		to = originaldraft.ToEmail
	}

	originaldraft.Title = title
	originaldraft.ToEmail = to
	originaldraft.Body = body

	err = server.db.UpdateDraftSentMessage(*originaldraft)
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

	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	basemsg, err := server.db.GetMessageFromReplyTo(id)
	if basemsg.Type == "sent" {
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
	} else {
		message, err := server.db.GetReceivedMessage(id)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve Received message from database", Error: err.Error(), Success: false}, http.StatusNotFound)
			return
		}

		if message.ToEmail != from {
			WriteJSON(w, Response{Data: "You don't have permission to delete this message", Success: false}, http.StatusForbidden)
			return
		}

		server.db.DeleteReceivedMessage(id)
	}
	server.db.DeleteMessage(id)

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}

func (server *httpImpl) RetrieveMessageFromRemoteServer(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := server.security.CheckUser(r)
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
	resp, err := req.Get(msg.URI)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to make a request", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	code := resp.Response().StatusCode
	bodystring := resp.String()

	if code != http.StatusOK {
		WriteJSON(w, Response{Data: bodystring, Success: false, Error: "Request failed on server side"}, code)
		return
	}
	server.logger.Debug(bodystring)

	// Let's manipulate string to hide URLs to attachments
	// TLDR: Some -advanced- HIGH TECH manipulation
	var j ReceivedMessageResponse
	err = json.Unmarshal(resp.Bytes(), &j)
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

	basemsg, err := server.db.GetMessageFromReplyTo(id)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Could not retrieve base message from database", Success: false}, http.StatusInternalServerError)
		return
	}
	replies, err := server.db.GetReplies(*basemsg, username)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Could not retrieve replies from database", Success: false}, http.StatusInternalServerError)
		return
	}
	j.Data.Replies = replies

	// We don't handle error, as it's not as important to mark as read message
	msg.IsRead = true
	err = server.db.UpdateReceivedMessage(*msg)
	server.logger.Debug(err)

	WriteJSON(w, j, http.StatusOK)
}

func (server *httpImpl) MarkReadUnread(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := server.security.CheckUser(r)
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
	msg.IsRead = !msg.IsRead
	err = server.db.UpdateReceivedMessage(*msg)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to mark as read/unread", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
}

func (server *httpImpl) GetSentMessageData(w http.ResponseWriter, r *http.Request) {
	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to parse ID", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	msg, err := server.db.GetSentMessage(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve Sent message from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	basemsg, err := server.db.GetMessageFromReplyTo(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve Base message from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	if msg.FromEmail != from {
		WriteForbiddenJWT(w, errors.New(""))
		return
	}

	attachments, err := server.db.GetAllAttachments(id)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve attachments", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	protocol := "http://"
	if server.config.HTTPSEnabled {
		protocol = "https://"
	}

	attachmentsList := make([]Attachment, 0)
	replies, err := server.db.GetReplies(*basemsg, from)

	for i := 0; i < len(attachments); i++ {
		a := attachments[i]
		attachment := Attachment{
			ID:       a.ID,
			Filename: a.OriginalName,
			URL:      fmt.Sprintf("%s%s/smtp2/attachment/get/%s/%s", protocol, server.config.HostURL, fmt.Sprint(id), fmt.Sprint(a.ID)),
		}
		attachmentsList = append(attachmentsList, attachment)
	}

	WriteJSON(w, Response{Data: ReceivedMessageData{
		ID:          id,
		Title:       msg.Title,
		Receiver:    msg.ToEmail,
		Body:        msg.Body,
		Sender:      msg.FromEmail,
		Attachments: attachmentsList,
		Replies:     replies,
	}, Success: true}, http.StatusOK)
}
