package helpers

import (
	"net/http"
	"net/url"
	"strings"
)

func Write(w http.ResponseWriter, text string, status int) (int, error) {
	w.WriteHeader(status)
	i, err := w.Write([]byte(text))
	return i, err
}

func GetDomainFromEmail(email string) string {
	strs := strings.Split(email, "@")
	tld := strs[len(strs)-1]
	return tld
}

func GetHostnameFromURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, err
}
