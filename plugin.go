package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"

	"github.com/gotify/plugin-api"
    "github.com/gorilla/websocket"
)

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
        Version: "1.0",
        Author: "Anh Bui",
		Name: "Gotify 2 Telegram",
        Description: "Telegram message fowarder for gotify",
		ModulePath: "https://github.com/anhbh310/gotify2telegram",

	}
}

// Plugin is the plugin instance
type Plugin struct {
    ws *websocket.Conn;
    msgHandler plugin.MessageHandler;
    chatid string;
    telegram_bot_token string;
    gotify_host string;
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

func (p *Plugin) send_msg_to_telegram(msg string) {
    step_size := 4090
    sending_message := ""
    for i:=0; i<len(msg); i+=step_size {
        if i+step_size < len(msg) {
			sending_message = msg[i : i+step_size]
		} else {
			sending_message = msg[i:len(msg)]
		}

        data := Payload{
        // Fill struct
            ChatID: p.chatid,
            Text: sending_message,
        }
        payloadBytes, err := json.Marshal(data)
        if err != nil {
            fmt.Println("Create json false")
            return
        }
        body := bytes.NewReader(payloadBytes)
        
        req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+ p.telegram_bot_token +"/sendMessage", body)
        if err != nil {
            fmt.Println("Create request false")
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

func (p *Plugin) connect_websocket() {
    for {
        ws, _, err := websocket.DefaultDialer.Dial(p.gotify_host, nil)
        if err == nil {
            p.ws = ws
            break
        }
        fmt.Printf("Cannot connect to websocket: %v\n", err)
        time.Sleep(5)
    }
}

func (p *Plugin) get_websocket_msg(url string, token string) {
    p.gotify_host = url + "/stream?token=" + token
    p.chatid = os.Getenv("TELEGRAM_CHAT_ID")
    p.telegram_bot_token = os.Getenv("TELEGRAM_BOT_TOKEN")

    go p.connect_websocket()

    for {
        msg := &GotifyMessage{}
        if p.ws == nil {
            time.Sleep(3)
            continue
        }
        err := p.ws.ReadJSON(msg)
        if err != nil {
            fmt.Printf("Error while reading websocket: %v\n", err)
            p.connect_websocket()
            continue
        }
        p.send_msg_to_telegram(msg.Date + "\n" + msg.Title + "\n\n" + msg.Message)
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