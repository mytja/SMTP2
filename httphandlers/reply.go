package httphandlers

import (
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

// TODO: Disable anyone with JWT to reply.
func NewReplyHandler(w http.ResponseWriter, r *http.Request) {
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

	replytoid, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}

	replytomsg, err := sql.DB.GetMessageFromReplyTo(replytoid)
	if err != nil {
		helpers.Write(w, "Failed retrieving original message", http.StatusInternalServerError)
		return
	}

	var originalid int
	if replytomsg.OriginalID == -1 {
		originalid = replytomsg.ID
	} else {
		originalid = replytomsg.OriginalID
	}

	id := sql.DB.GetLastMessageID()
	basemsg := objects.NewMessage(id, originalid, -1, replytomsg.ReplyPass, replytomsg.ReplyID, "sent")
	err = sql.DB.CommitMessage(basemsg)
	if err != nil {
		helpers.Write(w, "Failed while committing message base", http.StatusInternalServerError)
		return
	}

	// Here we get either SentMessage or ReceivedMessage
	var to string
	if replytomsg.Type == "sent" {
		message, err := sql.DB.GetSentMessage(replytoid)
		if err != nil {
			helpers.Write(w, "Failed to retrieve Sent message.", http.StatusInternalServerError)
			return
		}
		to = message.ToEmail
	} else {
		message, err := sql.DB.GetReceivedMessage(replytoid)
		if err != nil {
			helpers.Write(w, "Failed to retrieve Received message.", http.StatusInternalServerError)
			return
		}
		to = message.FromEmail
	}

	// Generate random password
	pass, err := security.GenerateRandomString(50)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reply := sql.NewSentMessage(title, to, from, body, pass)
	reply.ID = id
	fmt.Println(reply)
	err = sql.DB.CommitSentMessage(reply)
	if err != nil {
		helpers.Write(w, "Failed to commit Sent message", http.StatusInternalServerError)
		return
	}

	// Now let's send a request to a recipient email server
	domain := helpers.GetDomainFromEmail(to)
	fmt.Println(domain)

	protocol := "http://"
	if constants.ForceHttps {
		protocol = "https://"
	}
	reqdom := protocol + domain + "/smtp2/message/receive"
	req, err := http.NewRequest("POST", reqdom, strings.NewReader(""))
	req.Header.Set("Title", title)
	req.Header.Set("To", to)
	req.Header.Set("From", from)
	req.Header.Set("ServerPass", pass)
	req.Header.Set("ReplyPass", replytomsg.ReplyPass)
	req.Header.Set("ReplyID", replytomsg.ReplyID)
	req.Header.Set("ServerID", fmt.Sprint(id))
	req.Header.Set(
		"URI",
		protocol+helpers.GetDomainFromEmail(from)+"/smtp2/message/get/"+fmt.Sprint(id)+"?pass="+pass,
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
		helpers.Write(w, helpers.BytearrayToString(body2), http.StatusNotAcceptable)
		return
	}
	fmt.Println(req.Header.Get("URI"))
	fmt.Println(reqdom)
	helpers.Write(w, "Unknown error: "+fmt.Sprint(res.StatusCode)+" - "+helpers.BytearrayToString(body2), http.StatusInternalServerError)
}
