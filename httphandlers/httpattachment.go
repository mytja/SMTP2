package httphandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/imroc/req"
	"github.com/mytja/SMTP2/helpers"
	"github.com/mytja/SMTP2/security"
	"github.com/mytja/SMTP2/sql"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func (server *httpImpl) UploadFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve FormFile", Success: false}, http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	idint, err := strconv.Atoi(id)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "ID isn't a integer", Success: false}, http.StatusBadRequest)
		return
	}

	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	message, err := server.db.GetSentMessage(idint)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve message", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	if message.FromEmail != from {
		WriteJSON(w, Response{Data: "You didn't create this message...", Success: false}, http.StatusForbidden)
		return
	}

	pass, err := security.GenerateRandomString(80)
	if err != nil {
		WriteJSON(
			w,
			Response{
				Data:    "Failed to generate random password. This is completely server's fault",
				Error:   err.Error(),
				Success: false,
			},
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
		server.logger.Debug("Creating attachments directory")
		err := os.Mkdir("attachments", 0755)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to create attachments directory", Success: false}, http.StatusInternalServerError)
			return
		}
	}

	if _, err := os.Stat("attachments/" + id + "/"); errors.Is(err, os.ErrNotExist) {
		server.logger.Debug("Creating message directory")
		err := os.Mkdir("attachments/"+id, 0755)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to create message directory", Success: false}, http.StatusInternalServerError)
			return
		}
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to open file", Success: false}, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(f, file)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to copy attachment to file", Success: false}, http.StatusInternalServerError)
		return
	}

	defer f.Close()

	err = server.db.CommitAttachment(newattachment)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit attachment to database", Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: lastattachmentid, Success: true}, http.StatusCreated)
}

func (server *httpImpl) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		WriteJSON(w, Response{Data: "Message ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		WriteJSON(w, Response{Data: "Attachment ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	sentmsg, err := server.db.GetSentMessage(midint)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve sent message from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	if sentmsg.FromEmail != from {
		WriteJSON(w, Response{Data: "Forbidden", Success: false}, http.StatusForbidden)
		return
	}

	attachment, err := server.db.GetAttachment(midint, aidint)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve attachment from database", Success: false}, http.StatusInternalServerError)
		return
	}

	err = os.Remove(attachment.Filename)
	if err != nil {
		WriteJSON(w, Response{Data: "Error while trying to remove file", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	err = server.db.DeleteAttachment(midint, aidint)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to delete attachment from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "Successfully deleted following file.", Success: true}, http.StatusOK)
}

// GetAttachment Sender retrieves attachment on its own server
func (server *httpImpl) GetAttachment(w http.ResponseWriter, r *http.Request) {
	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		WriteJSON(w, Response{Data: "Message ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		WriteJSON(w, Response{Data: "Attachment ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	sentmsg, err := server.db.GetSentMessage(midint)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve sent message from database", Success: false}, http.StatusInternalServerError)
		return
	}

	if sentmsg.FromEmail != from {
		WriteJSON(w, Response{Data: "Forbidden", Success: false}, http.StatusForbidden)
		return
	}

	attachment, err := server.db.GetAttachment(midint, aidint)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve attachment from database", Success: false}, http.StatusInternalServerError)
		return
	}

	Openfile, err := os.Open(attachment.Filename)
	defer Openfile.Close()
	if err != nil {
		WriteJSON(w, Response{Data: "File not found.", Error: err.Error(), Success: false}, http.StatusNotFound)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	_, err = Openfile.Read(FileHeader)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to read file", Success: false}, http.StatusInternalServerError)
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
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, Openfile)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed writing file to writer.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
}

// RetrieveAttachment Recipient retrieves attachment from sender's server
func (server *httpImpl) RetrieveAttachment(w http.ResponseWriter, r *http.Request) {
	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		WriteJSON(w, Response{Data: "Message ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		WriteJSON(w, Response{Data: "Attachment ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	pass := r.URL.Query().Get("pass")

	attachment, err := server.db.GetAttachment(midint, aidint)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	if attachment.AttachmentPass != pass {
		WriteJSON(w, Response{Data: "Attachment password isn't matching to one saved in database", Success: false}, http.StatusForbidden)
		return
	}

	Openfile, err := os.Open(attachment.Filename)
	defer Openfile.Close()
	if err != nil {
		WriteJSON(w, Response{Data: "File not found.", Error: err.Error(), Success: false}, http.StatusNotFound)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	_, err = Openfile.Read(FileHeader)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
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
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, Openfile)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed writing file to writer.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
}

func (server *httpImpl) RetrieveAttachmentFromRemoteServer(w http.ResponseWriter, r *http.Request) {
	skipav := r.URL.Query().Get("skipav")
	ok, from, err := server.security.CheckUser(r)
	if err != nil || !ok {
		WriteForbiddenJWT(w, err)
		return
	}

	mid := mux.Vars(r)["mid"]
	midint, err := strconv.Atoi(mid)
	if err != nil {
		WriteJSON(w, Response{Data: "Message ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	aid := mux.Vars(r)["aid"]
	aidint, err := strconv.Atoi(aid)
	if err != nil {
		WriteJSON(w, Response{Data: "Attachment ID isn't a proper integer", Error: err.Error(), Success: false}, http.StatusBadRequest)
		return
	}

	message, err := server.db.GetReceivedMessage(midint)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to retrieve SentMessage from database", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	if message.ToEmail != from {
		WriteJSON(w, Response{Data: "This message doesn't belong to you", Success: false}, http.StatusForbidden)
		return
	}

	resp, err := req.Get(message.URI)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to make a request", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	bodystring := resp.String()
	code := resp.Response().StatusCode
	if code != http.StatusOK {
		WriteJSON(w, Response{Data: bodystring, Success: false}, code)
		return
	}
	var j MessagePayload
	err = json.Unmarshal(resp.Bytes(), &j)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to unmarshal response body", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	var attachments = j.Data.Attachments
	var url = ""
	for i := 0; i < len(attachments); i++ {
		attachment := attachments[i]
		if attachment.ID == aidint {
			url = attachment.URL
		}
	}
	if url == "" {
		WriteJSON(w, Response{Data: "Could not find attachment with following ID", Success: false}, http.StatusNotFound)
		return
	}
	att, err := req.Get(url)
	if err != nil {
		WriteJSON(w, Response{Data: "Failed to make a request", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	status := att.Response().StatusCode
	if status != 200 {
		WriteJSON(w, Response{Error: att.String(), Success: false}, status)
		return
	}

	contents := att.Bytes()

	if skipav != "1" {
		// Here goes AV analysis
		var analysisresult AttachmentAnalysis
		server.logger.Info(server.config.AV_URL)
		file := req.FileUpload{File: ioutil.NopCloser(bytes.NewReader(contents)), FieldName: "FILES"}
		server.logger.Info("Created file")
		req.Debug = true
		avreq, err := req.Post(server.config.AV_URL, file)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		server.logger.Info("Made request")
		avbody, err := avreq.ToBytes()
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to read response body from AV scan", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		server.logger.Info("AV request dumped to bytes")
		err = json.Unmarshal(avbody, &analysisresult)
		if err != nil {
			WriteJSON(w, Response{Data: "Failed to unmarshal AV response", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		server.logger.Info("AV request unmarshaled")
		if !analysisresult.Success {
			WriteJSON(w, Response{Data: fmt.Sprint(string(avbody), " - ", analysisresult), Success: false, Error: "AV scan failed"}, http.StatusBadGateway)
			return
		}
		if analysisresult.Data.Result[0].IsInfected {
			WriteJSON(w, Response{Data: "This file is infected with malware", Error: analysisresult.Data.Result[0].Viruses[0], Success: false}, http.StatusInternalServerError)
			return
		}
		server.logger.Info("AV request done", analysisresult)
	}

	_, err = io.Copy(w, ioutil.NopCloser(bytes.NewReader(contents)))
	if err != nil {
		WriteJSON(w, Response{Data: "Failed writing file to writer.", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	server.logger.Info("Wrote to writer")
}
