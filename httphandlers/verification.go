package httphandlers

import (
	"fmt"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/sql"
	"net/http"
	"strconv"
)

func MessageVerificationHandlers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id, err := strconv.Atoi(q.Get("id"))
	if err != nil {
		fmt.Println(err)
		helpers.Write(w, "FAIL", http.StatusInternalServerError)
		return
	}
	fmt.Println(id, q.Get("pass"))
	pass := q.Get("pass")
	sentmsg, err := sql.DB.GetSentMessage(id)
	if err != nil {
		fmt.Println(err)
		helpers.Write(w, "FAIL", http.StatusForbidden)
		return
	}
	if sentmsg.Pass != pass {
		helpers.Write(w, "Failed to verify Message password", http.StatusForbidden)
		return
	}
	helpers.Write(w, "OK", http.StatusOK)
}
