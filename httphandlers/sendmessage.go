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
	"strings"
)

func NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	to := r.FormValue("To")
	body := r.FormValue("Body")
	if !strings.Contains(to, "@") {
		helpers.Write(w, "Invalid To address", http.StatusBadRequest)
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
	id := sql.DB.GetLastMessageID()
	basemsg := objects.NewMessage(id, -1, -1, replyPass, replyID, "sent")
	err = sql.DB.CommitMessage(basemsg)
	if err != nil {
		helpers.Write(w, fmt.Sprint("Error while committing to database: ", err.Error()), http.StatusInternalServerError)
		return
	}
	msg := sql.NewSentMessage(title, to, from, body, pass)
	msg.ID = id
	fmt.Println(msg.ID)

	// Now let's send a request to a recipient email server
	domain := helpers.GetDomainFromEmail(msg.ToEmail)
	fmt.Println(domain)

	protocol := "http://"
	if constants.ForceHttps {
		protocol = "https://"
	}
	reqdom := protocol + domain + "/smtp2/message/receive"
	req, err := http.NewRequest("POST", reqdom, strings.NewReader(""))
	req.Header.Set("Title", msg.Title)
	req.Header.Set("To", msg.ToEmail)
	req.Header.Set("From", msg.FromEmail)
	req.Header.Set("ServerPass", msg.Pass)
	req.Header.Set("ReplyPass", basemsg.ReplyPass)
	req.Header.Set("ReplyID", basemsg.ReplyID)

	// We have to commit a message before we send a request
	err = sql.DB.CommitSentMessage(msg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idstring := fmt.Sprint(id)
	fmt.Println("ID2: ", idstring)

	req.Header.Set("ServerID", fmt.Sprint(idstring))
	req.Header.Set(
		"URI",
		protocol+helpers.GetDomainFromEmail(msg.FromEmail)+"/smtp2/message/get/"+fmt.Sprint(id)+"?pass="+msg.Pass,
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
	if res.StatusCode == http.StatusNotAcceptable {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			helpers.Write(w, "Error while reading request body", http.StatusInternalServerError)
			return
		}
		helpers.Write(w, helpers.BytearrayToString(body), http.StatusNotAcceptable)
	}
	fmt.Println(req.Header.Get("URI"))
	fmt.Println(reqdom)
	helpers.Write(w, "Unknown error: "+fmt.Sprint(res.StatusCode), http.StatusInternalServerError)
}
