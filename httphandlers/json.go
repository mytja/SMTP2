package httphandlers

import (
	"encoding/json"
	"net/http"
)

type MessageDataPayload struct {
	ID          int          `json:"ID"`
	Title       string       `json:"Title"`
	Receiver    string       `json:"Receiver"`
	Sender      string       `json:"Sender"`
	Body        string       `json:"Body"`
	Attachments []Attachment `json:"Attachments"`
}

type MessagePayload struct {
	Data MessageDataPayload `json:"data"`
}

type ServerInfo struct {
	HasHTTPS bool `json:"hasHTTPS"`
}

type AttachmentAnalysisResult struct {
	Name       string   `json:"name"`
	IsInfected bool     `json:"is_infected"`
	Viruses    []string `json:"viruses"`
}

type AttachmentAnalysisData struct {
	Result []AttachmentAnalysisResult `json:"result"`
}

type AttachmentAnalysis struct {
	Response
	Data AttachmentAnalysisData `json:"data"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

type MessageData struct {
	ID       int    `json:"ID"`
	ServerID int    `json:"ServerID"`
	Title    string `json:"Title"`
	URI      string `json:"URI"`
	Receiver string `json:"Receiver"`
	Sender   string `json:"Sender"`
	IsRead   bool   `json:"IsRead"`
}

type MessageDataResponse struct {
	Response
	Data MessageData `json:"data"`
}

type InboxDataResponse struct {
	Response
	Data []MessageData `json:"data"`
}

type ReceivedMessageData struct {
	ID          int          `json:"ID"`
	ServerID    int          `json:"ServerID"`
	Title       string       `json:"Title"`
	Receiver    string       `json:"Receiver"`
	Sender      string       `json:"Sender"`
	Attachments []Attachment `json:"Attachments"`
	Body        string       `json:"Body"`
}

type ReceivedMessageResponse struct {
	Response
	Data ReceivedMessageData `json:"data"`
}

type Attachment struct {
	ID       int    `json:"ID"`
	Filename string `json:"Filename"`
	URL      string `json:"URL"`
}

func DumpJSON(jsonstruct interface{}) []byte {
	marshal, _ := json.Marshal(jsonstruct)
	return marshal
}

func WriteJSON(w http.ResponseWriter, jsonstruct interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(DumpJSON(jsonstruct))
}

func WriteForbiddenJWT(w http.ResponseWriter, err error) {
	w.WriteHeader(403)
	w.Header().Set("Content-Type", "application/json")
	w.Write(DumpJSON(Response{Success: false, Data: "Forbidden", Error: err.Error()}))
}
