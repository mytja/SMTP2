package httphandlers

import (
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/objects"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"net/http"
)

func (server *httpImpl) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("Email")
	pass := r.FormValue("Pass")
	// Check if password is valid
	var user objects.User
	// TODO: Tole je lahko bolje
	err := server.db.GetDB().Get(&user, "SELECT * FROM users WHERE email=$1", email)
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

func (server *httpImpl) NewUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("Email")
	pass := r.FormValue("Pass")
	if email == "" || pass == "" {
		helpers.Write(w, "Bad Request. A parameter isn't provided", http.StatusBadRequest)
		return
	}
	_, err := helpers.GetDomainFromEmail(email)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check if user is already in DB
	var user objects.User
	var userCreated = true
	// TODO: Vem da je tole lahko bolje
	err = server.db.GetDB().Get(&user, "SELECT * FROM users WHERE email=$1", email)
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

	err = server.db.NewUser(email, password)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, "Success", http.StatusCreated)
}
