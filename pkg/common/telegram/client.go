package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Client for Telegram Bot API
type Client struct {
	BotToken string
	ChatID   string
}

// NewClient creates a new Telegram client
func NewClient(botToken, chatID string) *Client {
	return &Client{
		BotToken: botToken,
		ChatID:   chatID,
	}
}

// Message represents a Telegram message
type Message struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendMessage sends a message to the configured Telegram chat
func (c *Client) SendMessage(text string) error {
	message := Message{
		ChatID:    c.ChatID,
		Text:      text,
		ParseMode: "HTML",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling Telegram message: %v", err)
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.BotToken)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram API returned non-200 status: %d", resp.StatusCode)
		return fmt.Errorf("telegram API call failed with status: %d", resp.StatusCode)
	}

	log.Printf("Telegram message sent successfully")
	return nil
}

// SendErrorNotification sends a formatted error notification to Telegram
func (c *Client) SendErrorNotification(gameID, clientID, playerID, betID string, settingsError, rngError error) error {
	text := fmt.Sprintf(`
🚨 <b>RTP & RNG Server Failure Alert</b> 🚨

<b>Game:</b> %s
<b>Client ID:</b> %s
<b>Player ID:</b> %s
<b>Bet ID:</b> %s
<b>Timestamp:</b> %s

<b>Errors:</b>
• Settings(RTP) API: %v
• RNG(Outcome) API: %v

<b>Status:</b> Both critical APIs failed - immediate attention required!
	`, gameID, clientID, playerID, betID, time.Now().Format("2006-01-02 15:04:05 UTC"), settingsError, rngError)

	return c.SendMessage(text)
}
