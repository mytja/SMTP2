package httphandlers

import (
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"net/http"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := helpers.Write(w, "Hello!\nSMTP2 is running", http.StatusOK)
	if err != nil {
		fmt.Println(err)
	}
}
