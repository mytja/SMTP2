package httphandlers

import (
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/security"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NewMessageHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("Title")
	to := r.FormValue("To")
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
	pass, err := security.GenerateRandomString(25)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msg := sql.NewSentMessage(title, to, from, body, pass)
	msg.ID = sql.DB.GetLastSentID()
	fmt.Println(msg.ID)

	// Now let's send a request to a recipient email server
	domain := helpers.GetDomainFromEmail(msg.ToEmail)
	fmt.Println(domain)

	reqdom := "http://" + domain + "/smtp2/message/receive"
	req, err := http.NewRequest("POST", reqdom, strings.NewReader(""))
	req.Header.Set("Title", msg.Title)
	req.Header.Set("To", msg.ToEmail)
	req.Header.Set("From", msg.FromEmail)
	req.Header.Set("ServerPass", msg.Pass)

	// We have to commit a message before we send a request
	id, err := sql.DB.CommitSentMessage(msg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idstring := fmt.Sprint(id)
	fmt.Println("ID2: ", idstring)

	req.Header.Set("ServerID", fmt.Sprint(idstring))
	req.Header.Set("URI", "http://"+helpers.GetDomainFromEmail(msg.FromEmail)+"/message/get/"+fmt.Sprint(id))

	time.Sleep(1 * time.Second)

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
	fmt.Println(req.Header.Get("URI"))
	fmt.Println(reqdom)
	helpers.Write(w, "Unknown error: "+fmt.Sprint(res.StatusCode), http.StatusInternalServerError)
}

func ReceiveMessageHandler(w http.ResponseWriter, r *http.Request) {
	q := r.Header
	title := q.Get("Title")
	uri := q.Get("URI")
	to := q.Get("To")
	from := q.Get("From")
	id := q.Get("ServerID")
	pass := q.Get("ServerPass")
	fmt.Println(id, title, uri, to, from)
	atoi, err := strconv.Atoi(id)
	if err != nil {
		helpers.Write(w, "ID isn't a valid integer", http.StatusBadRequest)
		return
	}
	msg := sql.NewReceivedMessage(title, uri, to, from, atoi, pass)
	verification, _ := security.VerifyMessage(msg)
	if !verification {
		helpers.Write(w, "Failed to verify message.", http.StatusForbidden)
		return
	}
	msg.ID = sql.DB.GetLastReceivedID()
	err = sql.DB.CommitReceivedMessages(msg)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "Created", http.StatusCreated)
}

func MessageVerificationHandlers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id, err := strconv.Atoi(q.Get("id"))
	if err != nil {
		fmt.Println(err)
		helpers.Write(w, "FAIL", http.StatusInternalServerError)
		return
	}
	fmt.Println(id, q.Get("pass"))
	_, err = sql.DB.GetSentMessage(id, q.Get("pass"))
	if err != nil {
		fmt.Println(err)
		helpers.Write(w, "FAIL", http.StatusForbidden)
		return
	}
	helpers.Write(w, "OK", http.StatusOK)
}
