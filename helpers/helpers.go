package helpers

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func Write(w http.ResponseWriter, text string, status int) (int, error) {
	w.WriteHeader(status)
	i, err := w.Write([]byte(text))
	return i, err
}

func GetDomainFromEmail(email string) (string, error) {
	strs := strings.Split(email, "@")
	if len(strs) > 2 {
		return "", errors.New("string contains too much @ symbols - invalid email address")
	} else if len(strs) == 1 {
		return "", errors.New("string isn't a valid email address - no @ symbol was found")
	}
	tld := strs[len(strs)-1]
	return tld, nil
}

func GetHostnameFromURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, err
}

func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}
