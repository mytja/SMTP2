package httphandlers

import (
	"encoding/json"
	"github.com/mytja/SMTP2/helpers"
	"net/http"
)

func (server *httpImpl) ServerInfo(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["hasHTTPS"] = server.config.HTTPSEnabled
	j, err := json.Marshal(m)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.Write(w, helpers.BytearrayToString(j), http.StatusOK)
}
