package httphandlers

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/security"
	crypto2 "github.com/mytja/SMTP2/security/crypto"
	"github.com/mytja/SMTP2/sql"
	"io"
	"net/http"
	"os"
	"strconv"
)

func (server *httpImpl) UploadFile(w http.ResponseWriter, r *http.Request) {
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

	message, err := server.db.GetSentMessage(idint)
	if err != nil {
		helpers.Write(w, "Failed to retrieve message", http.StatusInternalServerError)
		return
	}

	if message.FromEmail != from {
		helpers.Write(w, "You didn't create this message...", http.StatusForbidden)
		return
	}

	pass, err := security.GenerateRandomString(80)
	if err != nil {
		helpers.Write(
			w,
			"Failed to generate random password. This is completely server's fault",
			http.StatusInternalServerError,
		)
		return
	}

	lastattachmentid := server.db.GetLastAttachmentID()
	fileext := helpers.GetFileExtension(handler.Filename)
	filename := "attachments/" + id + "/" + fmt.Sprint(lastattachmentid) + fileext
	newattachment := sql.NewAttachment(lastattachmentid, idint, handler.Filename, filename, pass)

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

	err = server.db.CommitAttachment(newattachment)
	if err != nil {
		helpers.Write(w, "Failed to commit attachment to database", http.StatusInternalServerError)
		return
	}

	helpers.Write(w, handler.Filename+" has been successfully uploaded", http.StatusCreated)
}

func (server *httpImpl) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		helpers.Write(w, "Message ID isn't a proper integer", http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		helpers.Write(w, "Attachment ID isn't a proper integer", http.StatusBadRequest)
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

	sentmsg, err := server.db.GetSentMessage(midint)
	if err != nil {
		helpers.Write(w, "", http.StatusInternalServerError)
		return
	}

	if sentmsg.FromEmail != from {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	attachment, err := server.db.GetAttachment(midint, aidint)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.Remove(attachment.Filename)
	if err != nil {
		helpers.Write(w, fmt.Sprint("Error while trying to remove file", err.Error()), http.StatusInternalServerError)
		return
	}

	err = server.db.DeleteAttachment(midint, aidint)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.Write(w, "Successfully deleted following file.", http.StatusOK)
}

func (server *httpImpl) GetAttachment(w http.ResponseWriter, r *http.Request) {
	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		helpers.Write(w, "Message ID isn't a proper integer", http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		helpers.Write(w, "Attachment ID isn't a proper integer", http.StatusBadRequest)
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

	sentmsg, err := server.db.GetSentMessage(midint)
	if err != nil {
		helpers.Write(w, "", http.StatusInternalServerError)
		return
	}

	if sentmsg.FromEmail != from {
		helpers.Write(w, "Forbidden", http.StatusForbidden)
		return
	}

	attachment, err := server.db.GetAttachment(midint, aidint)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Openfile, err := os.Open(attachment.Filename)
	defer Openfile.Close()
	if err != nil {
		helpers.Write(w, "File not found.", http.StatusNotFound)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	_, err = Openfile.Read(FileHeader)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	headers := w.Header()
	headers.Set("Content-Disposition", "attachment; filename="+attachment.OriginalName)
	headers.Set("Content-Type", FileContentType)
	headers.Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	_, err = Openfile.Seek(0, 0)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, Openfile)
	if err != nil {
		helpers.Write(w, fmt.Sprint("Failed writing file to writer.", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (server *httpImpl) RetrieveAttachment(w http.ResponseWriter, r *http.Request) {
	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		helpers.Write(w, "Message ID isn't a proper integer", http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		helpers.Write(w, "Attachment ID isn't a proper integer", http.StatusBadRequest)
		return
	}

	pass := r.URL.Query().Get("pass")

	attachment, err := server.db.GetAttachment(midint, aidint)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if attachment.AttachmentPass != pass {
		helpers.Write(w, "Attachment password isn't matching to one saved in database", http.StatusForbidden)
		return
	}

	Openfile, err := os.Open(attachment.Filename)
	defer Openfile.Close()
	if err != nil {
		helpers.Write(w, "File not found.", http.StatusNotFound)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	_, err = Openfile.Read(FileHeader)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	headers := w.Header()
	headers.Set("Content-Disposition", "attachment; filename="+attachment.OriginalName)
	headers.Set("Content-Type", FileContentType)
	headers.Set("Content-Length", FileSize)
	headers.Set("X-Filename", attachment.OriginalName)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	_, err = Openfile.Seek(0, 0)
	if err != nil {
		helpers.Write(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, Openfile)
	if err != nil {
		helpers.Write(w, fmt.Sprint("Failed writing file to writer.", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}