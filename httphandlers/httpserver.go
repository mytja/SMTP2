package httphandlers

import (
	"net/http"
)

func (server *httpImpl) ServerInfo(w http.ResponseWriter, r *http.Request) {
	m := ServerInfo{
		HasHTTPS: server.config.HTTPSEnabled,
	}
	WriteJSON(w, m, http.StatusOK)
}
