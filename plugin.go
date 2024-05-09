package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"

	"github.com/gotify/plugin-api"
    "github.com/gorilla/websocket"
)

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		Name:       "Gotify 2 Telegram",
		ModulePath: "https://github.com/anhbh310/gotify2telegram",
		Author:     "Anh Bui",
	}
}

// Plugin is the plugin instance
type Plugin struct {
    msgHandler plugin.MessageHandler;
}

type GotifyMessage struct {
    Id uint32;
    Appid uint32;
    Message string;
    Title string;
    Priority uint32;
    Date string;
}

type Payload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}
func (p *Plugin) check_websocket_connection() error {
    return nil
}

func (p *Plugin) send_msg_to_telegram(chatid string, telegram_api_token string, msg string) {
    data := Payload{
    // fill struct
        ChatID: chatid,
        Text: msg,
    }
    payloadBytes, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Create json false")
    }
    body := bytes.NewReader(payloadBytes)
    
    req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+ telegram_api_token +"/sendMessage", body)
    if err != nil {
        // handle err
        fmt.Println("Create request false")
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        // handle err
        fmt.Println("Send request false")
    }
    defer resp.Body.Close()
}

func (p *Plugin) get_websocket_msg(url string, token string) {
    chatid := os.Getenv("TELEGRAM_CHAT_ID")
    telegram_api_token := os.Getenv("TELEGRAM_API_TOKEN")
    ws, _, err := websocket.DefaultDialer.Dial(url + "/stream?token=" + token, nil)
    if err != nil {
        fmt.Println(err)
    }
    defer ws.Close()

    for {
        msg := &GotifyMessage{}
        err := ws.ReadJSON(msg)
        if err != nil {
            return
        }
        p.send_msg_to_telegram(chatid, telegram_api_token, msg.Date + "\n" + msg.Title + "\n" + msg.Message)
    }
}

// SetMessageHandler implements plugin.Messenger
// Invoked during initialization
func (p *Plugin) SetMessageHandler(h plugin.MessageHandler) {
    p.msgHandler = h
}

func (p *Plugin) Enable() error {
    go p.get_websocket_msg(os.Getenv("GOTIFY_HOST"), os.Getenv("GOTIFY_CLIENT_TOKEN"))
    return nil
}

// Disable implements plugin.Plugin
func (p *Plugin) Disable() error {
    return nil
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
    return &Plugin{}
}

func main() {
    // panic("this should be built as go plugin")
    // For testing
    p := &Plugin{nil}
    p.get_websocket_msg(os.Getenv("GOTIFY_HOST"), os.Getenv("GOTIFY_CLIENT_TOKEN"))
}