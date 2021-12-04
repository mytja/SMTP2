package httphandlers

import (
	"encoding/json"
	"net/http"
)

type MessagePayload struct {
	ID          string              `json:"ID"`
	Title       string              `json:"Title"`
	Receiver    string              `json:"Receiver"`
	Sender      string              `json:"Sender"`
	Body        string              `json:"Body"`
	Attachments []map[string]string `json:"Attachments"`
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

func DumpJSON(jsonstruct interface{}) []byte {
	marshal, _ := json.Marshal(jsonstruct)
	return marshal
}

func WriteJSON(w http.ResponseWriter, jsonstruct interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write(DumpJSON(jsonstruct))
}

func WriteForbiddenJWT(w http.ResponseWriter, err error) {
	w.WriteHeader(403)
	w.Write(DumpJSON(Response{Success: false, Data: "Forbidden", Error: err.Error()}))
}
