package httphandlers

import (
	"github.com/mytja/SMTP2/helpers"
	"net/http"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	helpers.Write(w, "Hello!\nSMTP2 is running", http.StatusOK)
}
