package handlers

import "net/http"

func Write(w http.ResponseWriter, text string, status int) (int, error) {
	w.WriteHeader(status)
	i, err := w.Write([]byte(text))
	return i, err
}
