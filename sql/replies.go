package sql

import (
	"errors"
	"fmt"
)

type SentMessageJSON struct {
	ID        int    `json:"ID"`
	Title     string `json:"Title"`
	Body      string `json:"Body"`
	ToEmail   string `json:"To"`
	FromEmail string `json:"From"`
	Type      string `json:"Type"`
}

type ReceivedMessageJSON struct {
	ID        int    `json:"ID"`
	Title     string `json:"Title"`
	URI       string `json:"URI"`
	ToEmail   string `json:"To"`
	FromEmail string `json:"From"`
	Type      string `json:"Type"`
}

func (db *sqlImpl) GetReplies(originalMessage Message, user string) ([]interface{}, error) {
	var message []Message
	err := db.db.Select(
		&message,
		"SELECT * FROM messages WHERE id>$1 AND reply_id=$2 AND reply_pass=$3",
		originalMessage.ID, originalMessage.ReplyID, originalMessage.ReplyPass,
	)
	if err != nil {
		return nil, err
	}
	messagesMap := make([]interface{}, 0)
	for i := 0; i < len(message); i++ {
		msg := message[i]
		if msg.Type == "sent" {
			sentMessage, err := db.GetSentMessage(msg.ID)
			if err != nil {
				return nil, err
			}
			if sentMessage.FromEmail != user {
				return nil, errors.New("unauthenticated")
			}
			app := SentMessageJSON{
				ToEmail:   sentMessage.ToEmail,
				FromEmail: sentMessage.FromEmail,
				ID:        sentMessage.ID,
				Title:     sentMessage.Title,
				Body:      sentMessage.Body,
				Type:      "sent",
			}
			messagesMap = append(messagesMap, app)
		}
		if msg.Type == "received" {
			sentMessage, err := db.GetReceivedMessage(msg.ID)
			if err != nil {
				return nil, err
			}
			if sentMessage.ToEmail != user {
				return nil, errors.New("unauthenticated")
			}

			app := ReceivedMessageJSON{
				ToEmail:   sentMessage.ToEmail,
				FromEmail: sentMessage.FromEmail,
				ID:        sentMessage.ID,
				Title:     sentMessage.Title,
				URI:       "/smtp2/message/retrieve/" + fmt.Sprint(sentMessage.ID),
				Type:      "received",
			}
			messagesMap = append(messagesMap, app)
		}
	}
	return messagesMap, nil
}
