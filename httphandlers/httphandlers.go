package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/objects"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"net/http"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := helpers.Write(w, "Hello!\nSMTP2 is running", http.StatusOK)
	if err != nil {
		fmt.Println(err)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("Email")
	pass := r.FormValue("Pass")
	// Check if password is valid
	var user objects.User
	err := sql.DB.GetDB().Get(&user, "SELECT * FROM users WHERE email=$1", email)
	hashCorrect := crypto2.CheckHash(pass, user.Password)
	if !hashCorrect {
		helpers.Write(w, "Hashes don't match...", http.StatusForbidden)
		return
	}

	// Extract JWT
	jwt, err := crypto2.GetJWTFromUserPass(email, user.Password)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.Write(w, jwt, http.StatusOK)
}

func NewUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("Email")
	pass := r.FormValue("Pass")
	if email == "" || pass == "" {
		helpers.Write(w, "Bad Request. A parameter isn't provided", http.StatusBadRequest)
		return
	}
	// Check if user is already in DB
	var user objects.User
	var userCreated = true
	err := sql.DB.GetDB().Get(&user, "SELECT * FROM users WHERE email=$1", email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			userCreated = false
		} else {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if userCreated == true {
		helpers.Write(w, "User is already in database", http.StatusUnprocessableEntity)
		return
	}

	password, err := crypto2.HashPassword(pass)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = sql.DB.NewUser(email, password)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "Success", http.StatusCreated)
}

func GetMessageHandler(w http.ResponseWriter, r *http.Request) {

}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, username, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	} else if isAuth == false {
		helpers.Write(w, "Not authenticated", http.StatusForbidden)
		return
	}
	fmt.Println(username)
	inbox, err := sql.DB.GetInbox(username)
	if err != nil {
		fmt.Println(err)
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

		m = append(m, m1)
	}
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(m)
	_, err = helpers.Write(w, helpers.BytearrayToString(response), http.StatusOK)
	if err != nil {
		return
	}
}
