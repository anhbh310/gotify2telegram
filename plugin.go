package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/http/httputil"
    "os"
    "strings"
    "time"

    "github.com/gotify/plugin-api"
    "github.com/gorilla/websocket"
)

// GetGotifyPluginInfo returns gotify plugin info
func GetGotifyPluginInfo() plugin.Info {
    return plugin.Info{
        Version:     "1.2",
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
    chatid            string
    telegram_bot_token string
    gotify_host       string
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
    ChatID    string `json:"chat_id"`
    Text      string `json:"text"`
    ParseMode string `json:"parse_mode"`
}

// escapeMarkdownV2 escapes special characters for Telegram's MarkdownV2 format
func escapeMarkdownV2(text string) string {
    // Special characters that need to be escaped in MarkdownV2
    specialChars := []string{"_", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    
    // Don't escape characters inside code blocks
    parts := strings.Split(text, "```")
    for i := 0; i < len(parts); i++ {
        if i%2 == 0 { // Outside code block
            for _, char := range specialChars {
                parts[i] = strings.ReplaceAll(parts[i], char, "\\"+char)
            }
        }
    }
    
    return strings.Join(parts, "```")
}

// splitMessage splits a message into chunks while preserving code blocks
func (p *Plugin) splitMessage(msg string, maxSize int) []string {
    if len(msg) <= maxSize {
        return []string{msg}
    }

    var chunks []string
    lines := strings.Split(msg, "\n")
    currentChunk := ""
    inCodeBlock := false
    
    for _, line := range lines {
        // Check for code block markers
        if strings.HasPrefix(strings.TrimSpace(line), "```") {
            inCodeBlock = !inCodeBlock
        }
        
        // If adding this line would exceed the limit
        if len(currentChunk)+len(line)+1 > maxSize {
            // If we're in a code block, close it in current chunk and reopen in next
            if inCodeBlock {
                currentChunk += "```"
                chunks = append(chunks, currentChunk)
                currentChunk = "```\n" + line + "\n"
            } else {
                chunks = append(chunks, currentChunk)
                currentChunk = line + "\n"
            }
        } else {
            currentChunk += line + "\n"
        }
    }
    
    // Add the final chunk if there's content
    if len(currentChunk) > 0 {
        chunks = append(chunks, currentChunk)
    }
    
    // Trim chunks and ensure code blocks are properly closed
    for i := range chunks {
        chunks[i] = strings.TrimSpace(chunks[i])
        
        // Count backticks to check if code block is properly closed
        count := strings.Count(chunks[i], "```")
        if count%2 == 1 {
            chunks[i] += "\n```"
        }
    }
    
    return chunks
}

func (p *Plugin) send_msg_to_telegram(msg string) {
    // Telegram's maximum message length is 4096 characters
    // We use 3800 to leave room for formatting
    const maxMessageSize = 3800
    
    // Split the message into chunks
    messageChunks := p.splitMessage(msg, maxMessageSize)
    
    totalChunks := len(messageChunks)
    
    for i, chunk := range messageChunks {
        // Add part number for multi-part messages
        var messageText string
        if totalChunks > 1 {
            messageText = fmt.Sprintf("(%d/%d)\n%s", i+1, totalChunks, chunk)
        } else {
            messageText = chunk
        }

        // Escape markdown characters
        escapedText := escapeMarkdownV2(messageText)

        data := Payload{
            ChatID:    p.chatid,
            Text:      escapedText,
            ParseMode: "MarkdownV2", // Use MarkdownV2 instead of Markdown
        }
        
        payloadBytes, err := json.Marshal(data)
        if err != nil {
            p.debugLogger.Printf("Create JSON failed: %v\n", err)
            continue
        }
        
        body := bytes.NewBuffer(payloadBytes)
        backup_body := bytes.NewBuffer(payloadBytes)

        req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+p.telegram_bot_token+"/sendMessage", body)
        if err != nil {
            p.debugLogger.Printf("Create request failed: %v\n", err)
            continue
        }
        req.Header.Set("Content-Type", "application/json")

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            p.debugLogger.Printf("Send request failed: %v\n", err)
            continue
        }

        if resp.StatusCode == http.StatusOK {
            p.debugLogger.Printf("Part %d/%d was forwarded successfully to Telegram\n", i+1, totalChunks)
        } else {
            p.debugLogger.Println("============== Request ==============")
            pretty_print, err := httputil.DumpRequest(req, true)
            if err != nil {
                p.debugLogger.Printf("%v\n", err)
            }
            p.debugLogger.Printf(string(pretty_print))
            p.debugLogger.Printf("%v\n", backup_body)

            p.debugLogger.Println("============== Response ==============")
            bodyBytes, err := io.ReadAll(resp.Body)
            p.debugLogger.Printf("%v\n", string(bodyBytes))
        }

        resp.Body.Close()
        
        // Add a small delay between messages to avoid rate limiting
        if totalChunks > 1 && i < totalChunks-1 {
            time.Sleep(500 * time.Millisecond)
        }
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

func (p *Plugin) SetMessageHandler(h plugin.MessageHandler) {
    p.debugLogger = log.New(os.Stdout, "Gotify 2 Telegram: ", log.Lshortfile)
    p.msgHandler = h
}

func (p *Plugin) Enable() error {
    go p.get_websocket_msg(os.Getenv("GOTIFY_HOST"), os.Getenv("GOTIFY_CLIENT_TOKEN"))
    return nil
}

func (p *Plugin) Disable() error {
    if p.ws != nil {
        p.ws.Close()
    }
    return nil
}

func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
    return &Plugin{}
}

func main() {
    panic("this should be built as go plugin")
}