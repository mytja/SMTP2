package httphandlers

import (
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/objects"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"net/http"
)

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
