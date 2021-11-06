package httphandlers

import (
	"github.com/mytja/SMTP2/helpers"
	"net/http"
	"strconv"
)

func (server *httpImpl) MessageVerificationHandlers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id, err := strconv.Atoi(q.Get("id"))
	if err != nil {
		server.logger.Info(err)
		helpers.Write(w, "FAIL", http.StatusInternalServerError)
		return
	}
	server.logger.Info(id, q.Get("pass"))
	pass := q.Get("pass")
	sentmsg, err := server.db.GetSentMessage(id)
	if err != nil {
		server.logger.Info(err)
		helpers.Write(w, "FAIL", http.StatusForbidden)
		return
	}
	if sentmsg.Pass != pass {
		helpers.Write(w, "Failed to verify Message password", http.StatusForbidden)
		return
	}
	helpers.Write(w, "OK", http.StatusOK)
}
