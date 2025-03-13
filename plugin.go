package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		Version:     "1.0",
		Author:      "Anh Bui",
		Name:        "Gotify 2 Telegram",
		Description: "Telegram message fowarder for gotify",
		ModulePath:  "https://github.com/anhbh310/gotify2telegram",
	}
}

// Plugin is the plugin instance
type Plugin struct {
	ws                 *websocket.Conn
	msgHandler         plugin.MessageHandler
	debugLogger        *log.Logger
	chatid             string
	telegram_bot_token string
	gotify_host        string
}

type GotifyMessage struct {
	Id       uint32
	Appid    uint32
	Message  string
	Title    string
	Priority uint32
	Date     string
}

type Payload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func (p *Plugin) send_msg_to_telegram(msg string) {
	step_size := 4090
	sending_message := ""

	for i := 0; i < len(msg); i += step_size {
		if i+step_size < len(msg) {
			sending_message = msg[i : i+step_size]
		} else {
			sending_message = msg[i:]
		}

		data := Payload{
			// Fill struct
			ChatID: p.chatid,
			Text:   sending_message,
		}
		payloadBytes, err := json.Marshal(data)
		if err != nil {
			p.debugLogger.Println("Create json false")
			return
		}
		body := bytes.NewBuffer(payloadBytes)
		// For future debugging
		backup_body := bytes.NewBuffer(body.Bytes())

		req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+p.telegram_bot_token+"/sendMessage", body)
		if err != nil {
			p.debugLogger.Println("Create request false")
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			p.debugLogger.Printf("Send request false: %v\n", err)
			return
		}
		p.debugLogger.Println("HTTP request was sent successfully")

		if resp.StatusCode == http.StatusOK {
			p.debugLogger.Println("The message was forwarded successfully to Telegram")
		} else {
			// Log infor for debugging
			p.debugLogger.Println("============== Request ==============")
			pretty_print, err := httputil.DumpRequest(req, true)
			if err != nil {
				p.debugLogger.Printf("%v\n", err)
			}
			p.debugLogger.Print(string(pretty_print))
			p.debugLogger.Printf("%v\n", backup_body)

			p.debugLogger.Println("============== Response ==============")
			bodyBytes, _ := io.ReadAll(resp.Body)
			p.debugLogger.Printf("%v\n", string(bodyBytes))

		}

		defer resp.Body.Close()
	}
}

func (p *Plugin) connect_websocket() {
	for {
		ws, _, err := websocket.DefaultDialer.Dial(p.gotify_host, nil)
		if err == nil {
			p.ws = ws
			break
		}
		p.debugLogger.Printf("Cannot connect to websocket: %v\n", err)
		time.Sleep(5 * time.Second)
	}
	p.debugLogger.Println("WebSocket connected successfully, ready for forwarding")
}

func (p *Plugin) get_websocket_msg(url string, token string) {
	p.gotify_host = url + "/stream?token=" + token
	p.chatid = os.Getenv("TELEGRAM_CHAT_ID")
	p.debugLogger.Printf("chatid: %v\n", p.chatid)
	p.telegram_bot_token = os.Getenv("TELEGRAM_BOT_TOKEN")
	p.debugLogger.Printf("Bot token: %v\n", p.telegram_bot_token)

	go p.connect_websocket()

	for {
		msg := &GotifyMessage{}
		if p.ws == nil {
			time.Sleep(3 * time.Second)
			continue
		}
		err := p.ws.ReadJSON(msg)
		if err != nil {
			p.debugLogger.Printf("Error while reading websocket: %v\n", err)
			p.connect_websocket()
			continue
		}
		p.send_msg_to_telegram(msg.Date + "\n" + msg.Title + "\n\n" + msg.Message)
	}
}

// SetMessageHandler implements plugin.Messenger
// Invoked during initialization
func (p *Plugin) SetMessageHandler(h plugin.MessageHandler) {
	p.debugLogger = log.New(os.Stdout, "Gotify 2 Telegram: ", log.Lshortfile)
	p.msgHandler = h
}

func (p *Plugin) Enable() error {
	go p.get_websocket_msg("ws://localhost", os.Getenv("GOTIFY_CLIENT_TOKEN"))
	return nil
}

// Disable implements plugin.Plugin
func (p *Plugin) Disable() error {
	if p.ws != nil {
		p.ws.Close()
	}
	return nil
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{}
}

func main() {
	panic("this should be built as go plugin")
	// For testing
	// p := &Plugin{nil, nil, "", "", ""}
	// p.get_websocket_msg(os.Getenv("GOTIFY_HOST"), os.Getenv("GOTIFY_CLIENT_TOKEN"))
}
