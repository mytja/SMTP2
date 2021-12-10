package httphandlers

import (
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/security"
	"net/http"
)

func (server *httpImpl) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("Email")
	pass := r.FormValue("Pass")
	// Check if password is valid
	user, err := server.db.GetUserByEmail(email)
	hashCorrect := security.CheckHash(pass, user.Password)
	if !hashCorrect {
		WriteJSON(w, Response{Data: "Hashes don't match...", Success: false}, http.StatusForbidden)
		return
	}

	// Extract JWT
	jwt, err := security.GetJWTFromUserPass(email)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: jwt, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("Email")
	pass := r.FormValue("Pass")
	if email == "" || pass == "" {
		WriteJSON(w, Response{Data: "Bad Request. A parameter isn't provided", Success: false}, http.StatusBadRequest)
		return
	}
	domain, err := helpers.GetDomainFromEmail(email)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to extract domain from email", Success: false}, http.StatusBadRequest)
		return
	}
	if server.config.HostURL != domain {
		WriteJSON(w, Response{Data: "This server doesn't host your domain", Success: false}, http.StatusForbidden)
		return
	}
	// Check if user is already in DB
	var userCreated = true
	_, err = server.db.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			userCreated = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Data: "Could not retrieve user from database", Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if userCreated == true {
		WriteJSON(w, Response{Data: "User is already in database", Success: false}, http.StatusUnprocessableEntity)
		return
	}

	password, err := security.HashPassword(pass)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to hash your password", Success: false}, http.StatusInternalServerError)
		return
	}

	err = server.db.NewUser(email, password)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit new user to database", Success: false}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, Response{Data: "Success", Success: true}, http.StatusCreated)
}
