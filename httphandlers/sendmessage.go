package httphandlers

import (
	"encoding/json"
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

	fromemail, err := helpers.GetDomainFromEmail(from)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fromemail != server.config.HostURL {
		helpers.Write(w, "This server doesn't hold your domain.", http.StatusForbidden)
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
	}

	if !strings.Contains(to, "@") {
		helpers.Write(w, "Invalid To address", http.StatusBadRequest)
		return
	}

	var responseMap = make([]map[string]interface{}, 0)

	tolist := strings.Split(to, ";")
	for i := 0; i < len(tolist); i++ {
		to = tolist[i]

		// Generate different passwords for different recipients
		pass, err := security.GenerateRandomString(80)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
		replyPass, err := security.GenerateRandomString(80)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
		replyID, err := security.GenerateRandomString(80)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mvppass, err := security.GenerateRandomString(80)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := server.db.GetLastMessageID()

		var result = make(map[string]interface{})

		todomain, err := helpers.GetDomainFromEmail(to)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := http.Get("http://" + todomain + "/smtp2/server/info")
		if err != nil {
			result["Body"] = "Remote server isn't available at the moment. Failed to send message."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}
		reqbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result["Body"] = "Failed to read response body."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}
		resjson := make(map[string]interface{})
		err = json.Unmarshal(reqbody, &resjson)
		if err != nil {
			result["Body"] = "Failed to unmarshal server's response."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}
		protocol := "http://"
		if resjson["hasHTTPS"] == true {
			protocol = "https://"
		}

		basemsg := objects.NewMessage(id, originalid, serverid, replyPass, replyID, "sent", false)
		err = server.db.CommitMessage(basemsg)
		if err != nil {
			result["Body"] = "Error while committing base Message to database."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}

		msg := sql.NewSentMessage(id, title, to, from, body, pass, mvppass)

		// Now let's send a request to a recipient email server

		idstring := fmt.Sprint(id)
		server.logger.Infow("requesting to "+to, "basemsg", basemsg, "msg", msg, "todomain", todomain, "id", idstring)

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
		req.Header.Set("MVPPass", fmt.Sprint(mvppass))
		req.Header.Set(
			"URI",
			urlprotocol+server.config.HostURL+"/smtp2/message/get/"+fmt.Sprint(id)+"?pass="+msg.Pass,
		)

		// We have to commit a message before we send a request
		err = server.db.CommitSentMessage(msg)
		if err != nil {
			result["Body"] = "Error while committing Sent message to database."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			result["Body"] = "Remote server isn't available at the moment. Failed to send message."
			result["To"] = to
			result["StatusCode"] = -1
			responseMap = append(responseMap, result)
			break
		}

		server.logger.Infow("requesting to "+to, "url", req.Header.Get("URI"), "domain", reqdom)

		body3, _ := ioutil.ReadAll(res.Body)
		result["Body"] = helpers.BytearrayToString(body3)
		result["To"] = to
		result["StatusCode"] = res.StatusCode
		responseMap = append(responseMap, result)

		server.logger.Info(helpers.BytearrayToString(body3))
		if res.StatusCode == http.StatusCreated {
		} else if res.StatusCode == http.StatusNotAcceptable || constants.EnableDeletingOnUnknownError {
			server.db.DeleteMessage(basemsg.ID)
			server.db.DeleteSentMessage(msg.ID)
		}
	}
	if iscreatedfromdraft == true {
		// Delete draft message, as it isn't used anymore
		server.logger.Infow("deleting draft message", "id", draftid)
		server.db.DeleteMessage(draftid)
		server.db.DeleteSentMessage(draftid)
	}
	marshal, _ := json.Marshal(responseMap)
	helpers.Write(w, helpers.BytearrayToString(marshal), http.StatusCreated)
}
