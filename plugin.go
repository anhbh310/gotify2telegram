package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		Version:     "1.1",
		Author:      "Anh Bui & Leko",
		Name:        "Gotify 2 Telegram",
		Description: "Telegram message fowarder for gotify",
		ModulePath:  "https://github.com/anhbh310/gotify2telegram",
	}
}

// Plugin is the plugin instance
type Plugin struct {
	config     *Config
	ws         *websocket.Conn
	msgHandler plugin.MessageHandler
	isEnabled  bool
}

type GotifyMessage struct {
	Id       uint32
	Appid    uint32
	Message  string
	Title    string
	Priority uint32
	Date     string
}

func (p *Plugin) connect_websocket(url string) {
	for {
		debug("connect_websocket:: Connecting ws.")
		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			ws.SetCloseHandler(func(code int, text string) error {
				p.ws = nil
				return fmt.Errorf("WebSocket Connection Closed. Resetting plugin.ws.")
			})
			p.ws = ws

			// Get Local Port
			if tcpConn, ok := ws.UnderlyingConn().(*net.TCPConn); ok {
				localAddr := tcpConn.LocalAddr().(*net.TCPAddr)
				debug("connect_websocket:: Connection Success. Port: %d", localAddr.Port)
			} else {
				debug("connect_websocket:: Connection Success. Underlying connection is not a TCP connection")
			}

			break
		}

		fmt.Printf("Cannot connect to websocket: %v\n", err)
		time.Sleep(5 * time.Second)
	}
}

func (p *Plugin) get_websocket_msg(url string) {
	ws_url := url + "/stream?token=" + p.config.GotifyClientToken
	p.connect_websocket(ws_url) // Initial connection setup

	for {
		if !p.isEnabled {
			break
		}

		if p.ws == nil {
			// If WebSocket is not initialized, attempt to reconnect
			time.Sleep(3 * time.Second)
			p.connect_websocket(ws_url)
		}

		msg := &GotifyMessage{}
		err := p.ws.ReadJSON(msg)
		if err != nil {
			fmt.Printf("Error while reading WebSocket: %v\n", err)
			p.ws.Close()
			continue
		}

		// Process the received message
		for _, subClient := range p.config.Clients {
			if subClient.AppId == int(msg.Appid) || subClient.AppId == -1 {
				debug("get_websocket_msg: AppId Matched! Sending to telegram...")
				send_msg_to_telegram(
					format_telegram_message(msg),
					subClient.Telegram.BotToken,
					subClient.Telegram.ChatId,
					subClient.Telegram.ThreadId,
				)
				break
			}
		}
	}
}

// SetMessageHandler implements plugin.Messenger
// Invoked during initialization
func (p *Plugin) SetMessageHandler(h plugin.MessageHandler) {
	p.msgHandler = h
}

func (p *Plugin) Enable() error {
	p.isEnabled = true
	go p.get_websocket_msg(p.config.GotifyHost)
	return nil
}

// Disable implements plugin.Plugin
func (p *Plugin) Disable() error {
	p.isEnabled = false
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
}
