package main

import (
	"fmt"
)

type Telegram struct {
	ChatId	 string	 `yaml:"chat_id"`
	BotToken string	 `yaml:"token"`
	ThreadId string	 `yaml:"thread_id"`
}

type SubClient struct {
	AppId		int		 `yaml:"app_id"`
	Telegram	Telegram `yaml:"telegram"`
}

// Config is user plugin configuration
type Config struct {
	Clients	   			[]SubClient	`yaml:"clients"`
	GotifyHost 			string		`yaml:"gotify_host"`
	GotifyClientToken	string	 	`yaml:"token"`
}

// DefaultConfig implements plugin.Configurer
func (c *Plugin) DefaultConfig() interface{} {
	return &Config{
		Clients: []SubClient{
			SubClient{
				AppId: 0,
				Telegram: Telegram{
					ChatId:	"-100123456789",
					BotToken: "YourBotTokenHere",
					ThreadId: "OptionalThreadIdHere",
				},
			},
		},
		GotifyHost: "ws://localhost:80",
		GotifyClientToken: "ExampleToken",
	}
}

// ValidateAndSetConfig implements plugin.Configurer
func (c *Plugin) ValidateAndSetConfig(config interface{}) error {
	newConfig := config.(*Config)

	if newConfig.GotifyClientToken == "ExampleToken" {
		return fmt.Errorf("gotify client token is required")
	}
	for i, client := range newConfig.Clients {
		if client.AppId == 0 {
			return fmt.Errorf("gotify app id is required for client %d", i)
		}
		if client.Telegram.BotToken == "" {
			return fmt.Errorf("telegram bot token is required for client %d", i)
		}
		if client.Telegram.ChatId == "" {
			return fmt.Errorf("telegram chat id is required for client %d", i)
		}
	}
  
	c.config = newConfig
	return nil
}