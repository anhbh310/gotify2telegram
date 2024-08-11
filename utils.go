package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func debug(s string, x ...interface{}) {
	log.Printf("Gotify2Telegram::"+s, x...)
}

func format_telegram_message(msg *GotifyMessage) string {
	// HTML Should be escaped here
	title := string(template.HTML("<b>" + template.HTMLEscapeString(msg.Title) + "</b>"))
	return fmt.Sprintf(
		"%s\n%s\n\nDate: %s",
		title,
		template.HTMLEscapeString(msg.Message),
		msg.Date,
	)
}

type Payload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ThreadId  string `json:"message_thread_id"`
	ParseMode string `json:"parse_mode"`
}

func send_msg_to_telegram(msg string, bot_token string, chat_id string, thread_id string) {
	step_size := 4090
	sending_message := ""
	for i := 0; i < len(msg); i += step_size {
		if i+step_size < len(msg) {
			sending_message = msg[i : i+step_size]
		} else {
			sending_message = msg[i:len(msg)]
		}

		data := Payload{
			ChatID:    chat_id,
			Text:      sending_message,
			ThreadId:  thread_id,
			ParseMode: "HTML",
		}
		payloadBytes, err := json.Marshal(data)
		if err != nil {
			log.Println("Create json false")
			return
		}
		body := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+bot_token+"/sendMessage", body)
		if err != nil {
			log.Println("Create request false")
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Send request false: %v\n", err)
			return
		}
		defer resp.Body.Close()
	}
}
