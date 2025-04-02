// internal/platform/settings/client.go
package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Client for game settings service
type Client struct {
	ServiceURL string
}

// NewClient creates a new settings client
func NewClient(serviceURL string) *Client {
	return &Client{
		ServiceURL: serviceURL,
	}
}

type Request struct {
	ClientID string `json:"client_id"`
	GameID   string `json:"game_id"`
	PlayerID string `json:"player_id"`
}

type Response struct {
	Data struct {
		GameBets string `json:"game_bets"`
		GameRTP  string `json:"game_rtp"`
		GameWins string `json:"game_wins"`
	} `json:"data"`
}

// GetRTP retrieves the RTP settings for a player
func (c *Client) GetRTP(clientID, gameID, playerID string) (float64, error) {
    reqBody, err := json.Marshal(Request{
        ClientID: clientID,
        GameID:   gameID,
        PlayerID: playerID,
    })
    if err != nil {
        log.Printf("Error marshaling settings request: %v", err)
        return 0, err
    }

    fmt.Printf("Settings request: %s\n", string(reqBody))

    resp, err := http.Post(c.ServiceURL, "application/json", bytes.NewBuffer(reqBody))
    if err != nil {
        log.Printf("Error calling settings API: %v", err)
        return 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("Settings API returned non-200 status: %d", resp.StatusCode)
        return 0, errors.New("Settings API call failed")
    }

    var settingsResp Response
    if err := json.NewDecoder(resp.Body).Decode(&settingsResp); err != nil {
        log.Printf("Error decoding settings response: %v", err)
        return 0, err
    }

    rtp, err := strconv.ParseFloat(settingsResp.Data.GameRTP, 64)
    if err != nil {
        log.Printf("Error parsing RTP value: %v", err)
        return 0, err
    }
    return rtp, nil
}