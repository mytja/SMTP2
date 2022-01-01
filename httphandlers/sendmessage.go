package httphandlers

import (
	"fmt"
	"github.com/imroc/req"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/security"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
	"strings"
)

type SentMessage struct {
	Title      string `json:"Title"`
	To         string `json:"To"`
	From       string `json:"From"`
	ServerPass string `json:"ServerPass"`
	ReplyPass  string `json:"ReplyPass"`
	ReplyID    string `json:"ReplyID"`
	OriginalID string `json:"OriginalID"`
	ServerID   string `json:"ServerID"`
	MVPPass    string `json:"MVPPass"`
	URI        string `json:"URI"`
}

func (server *httpImpl) NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	to := r.FormValue("To")
	body := r.FormValue("Body")
	usedraft := r.FormValue("DraftID")

	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	fromemail, err := helpers.GetDomainFromEmail(from)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to parse email address to get domain", Success: false}, http.StatusInternalServerError)
		return
	}
	if fromemail != server.config.HostURL && !server.config.SkipSameDomainCheck {
		WriteJSON(w, Response{Data: "This server doesn't hold your domain.", Success: false}, http.StatusForbidden)
		return
	}
	server.logger.Infow("config", "from", fromemail, "hosturl", server.config.HostURL)

	var iscreatedfromdraft = false
	var originalid = -1
	var serverid = -1
	var draftid int

	if usedraft != "" {
		draftid, err = strconv.Atoi(usedraft)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to parse provided DraftID to integer", Success: false}, http.StatusBadRequest)
			return
		}
		message, err := server.db.GetSentMessage(draftid)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve draft from database", Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		if message.FromEmail != from {
			WriteJSON(w, Response{Data: "You didn't create this draft...", Success: false}, http.StatusForbidden)
			return
		}
		basemsg, err := server.db.GetMessageFromReplyTo(draftid)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to retrieve draft base from database", Success: false, Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		if !basemsg.IsDraft {
			WriteJSON(w, Response{Data: "This isn't a draft anymore...", Success: false}, http.StatusBadRequest)
			return
		}
		iscreatedfromdraft = true
		serverid = basemsg.ServerID
		originalid = basemsg.OriginalID

		title = message.Title
		body = message.Body
		to = message.ToEmail
	}
	if !strings.Contains(to, "@") {
		WriteJSON(w, Response{Data: "Invalid To address", Success: false}, http.StatusBadRequest)
		return
	}

	var responseMap = make([]map[string]interface{}, 0)

	tolist := strings.Split(to, ";")
	for i := 0; i < len(tolist); i++ {
		to = tolist[i]

		// Generate different passwords for different recipients
		// TODO: Migrate this to break statements
		pass, err := security.GenerateRandomString(80)
		if err != nil {
			// This is a fatal error, thus we return and not just break
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to generate random password", Success: false}, http.StatusInternalServerError)
			return
		}
		replyPass, err := security.GenerateRandomString(80)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to generate random password", Success: false}, http.StatusInternalServerError)
			return
		}
		replyID, err := security.GenerateRandomString(80)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to generate random password", Success: false}, http.StatusInternalServerError)
			return
		}
		mvppass, err := security.GenerateRandomString(80)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to generate random password", Success: false}, http.StatusInternalServerError)
			return
		}

		id := server.db.GetLastMessageID()

		var result = make(map[string]interface{})

		// TODO: Migrate this to break statements
		todomain, err := helpers.GetDomainFromEmail(to)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to get domain from to email", Success: false}, http.StatusBadRequest)
			return
		}

		protocol, err := server.security.GetProtocolFromDomain(todomain)
		if err != nil {
			result["Body"] = "Remote server isn't available at the moment. Please try again later"
			result["To"] = to
			result["StatusCode"] = -1
			result["Error"] = err.Error()
			responseMap = append(responseMap, result)
			break
		}
		basemsg := sql.NewMessage(id, originalid, serverid, replyPass, replyID, "sent", false)
		err = server.db.CommitMessage(basemsg)
		if err != nil {
			result["Body"] = "Error while committing base Message to database."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}

		if iscreatedfromdraft == true {
			server.logger.Debug("migrating all attachments to this message")
			attachments, err := server.db.GetAllAttachments(draftid)
			if err != nil {
				result["Body"] = "Failed to retrieve attachments from database. " + err.Error()
				result["To"] = to
				result["StatusCode"] = -1
				responseMap = append(responseMap, result)
				break
			}
			for i := 0; i < len(attachments); i++ {
				attachment := attachments[i]
				// Generate new unique ID
				attachment.ID = server.db.GetLastAttachmentID()
				// Assign new message ID
				attachment.MessageID = id
				err := server.db.CommitAttachment(attachment)
				if err != nil {
					result["Body"] = "Failed to commit attachment from database. " + err.Error()
					result["To"] = to
					result["StatusCode"] = -1
					responseMap = append(responseMap, result)
					break
				}
			}
		}

		msg := sql.NewSentMessage(id, title, to, from, body, pass, mvppass)

		// Now let's send a request to a recipient email server

		idstring := fmt.Sprint(id)
		server.logger.Debugw("requesting to "+to, "basemsg", basemsg, "msg", msg, "todomain", todomain, "id", idstring)

		reqdom := protocol + todomain + "/smtp2/message/receive"

		urlprotocol := "http://"
		if server.config.HTTPSEnabled {
			urlprotocol = "https://"
		}

		// We have to commit a message before we send a request
		err = server.db.CommitSentMessage(msg)
		if err != nil {
			result["Body"] = "Error while committing Sent message to database."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}

		mailurl := urlprotocol + server.config.HostURL + "/smtp2/message/get/" + fmt.Sprint(id) + "?pass=" + msg.Pass

		headers := SentMessage{
			Title:      msg.Title,
			To:         msg.ToEmail,
			From:       msg.FromEmail,
			ServerPass: msg.Pass,
			ReplyPass:  basemsg.ReplyPass,
			ReplyID:    basemsg.ReplyID,
			OriginalID: "-1",
			ServerID:   idstring,
			MVPPass:    mvppass,
			URI:        mailurl,
		}

		res, err := req.Post(reqdom, req.HeaderFromStruct(headers))
		if err != nil {
			server.logger.Info(err)
			result["Body"] = "Remote server isn't available at the moment. Failed to send message."
			result["To"] = to
			result["StatusCode"] = -1
			result["Error"] = err.Error()
			responseMap = append(responseMap, result)
			break
		}

		server.logger.Debugw("requesting to "+to, "mail_url", mailurl, "domain", reqdom)

		code := res.Response().StatusCode

		result["Body"] = res.String()
		result["To"] = to
		result["StatusCode"] = code
		responseMap = append(responseMap, result)

		if code == http.StatusCreated {
			// Mogoče dodaj tukaj kej
		} else if code == http.StatusNotAcceptable || helpers.EnableDeletingOnUnknownError {
			server.db.DeleteMessage(basemsg.ID)
			server.db.DeleteSentMessage(msg.ID)
		}
	}
	if iscreatedfromdraft == true {
		// Delete draft message, as it isn't used anymore
		server.logger.Infow("deleting draft message", "id", draftid)
		server.db.DeleteMessage(draftid)
		server.db.DeleteSentMessage(draftid)
		attachments, err := server.db.GetAllAttachments(draftid)
		if err != nil {
			WriteJSON(
				w,
				Response{
					Error:   err.Error(),
					Data:    "Failed to retrieve all attachments from database. Your message was still sent",
					Success: false,
				},
				http.StatusInternalServerError,
			)
		}
		for i := 0; i < len(attachments); i++ {
			err := server.db.DeleteAttachment(draftid, attachments[0].ID)
			if err != nil {
				WriteJSON(
					w,
					Response{
						Data:    "Failed to delete old attachments from database. Your message was sent anyways",
						Error:   err.Error(),
						Success: false,
					},
					http.StatusInternalServerError,
				)
			}
		}
	}
	WriteJSON(w, Response{Data: responseMap, Success: true}, http.StatusCreated)
}
