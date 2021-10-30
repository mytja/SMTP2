package httphandlers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"io"
	"net/http"
	"os"
	"strconv"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
	}

	id := mux.Vars(r)["id"]
	idint, err := strconv.Atoi(id)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok, from, err := crypto2.CheckUser(r)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	message, err := sql.DB.GetSentMessage(idint)
	if err != nil {
		helpers.Write(w, "Failed to retrieve message", http.StatusInternalServerError)
		return
	}

	if message.FromEmail != from {
		helpers.Write(w, "You didn't create this message...", http.StatusForbidden)
		return
	}

	lastattachmentid := sql.DB.GetLastAttachmentID()
	fileext := helpers.GetFileExtension(handler.Filename)
	filename := "attachments/" + id + "/" + fmt.Sprint(lastattachmentid) + fileext
	newattachment := sql.NewAttachment(lastattachmentid, idint, handler.Filename, filename)

	defer file.Close()

	if _, err := os.Stat("attachments/"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Creating attachments directory")
		err := os.Mkdir("attachments", 0755)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if _, err := os.Stat("attachments/" + id + "/"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Creating message directory")
		err := os.Mkdir("attachments/"+id, 0755)
		if err != nil {
			helpers.Write(w, err.Error(), http.StatusInternalServerError)
		}
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(f, file)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer f.Close()

	err = sql.DB.CommitAttachment(newattachment)
	if err != nil {
		helpers.Write(w, "Failed to commit attachment to database", http.StatusInternalServerError)
		return
	}

	helpers.Write(w, handler.Filename+" has been successfully uploaded", http.StatusCreated)
}
