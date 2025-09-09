package rng

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/config"
	"github.com/google/uuid"
)

// ServiceState represents the health state of a service
type ServiceState int

const (
	HEALTHY ServiceState = iota
	DEGRADED
	FAILED
	RECOVERING
)

// ServiceEndpoint represents a single RNG service endpoint
type ServiceEndpoint struct {
	URL     string
	Timeout time.Duration
	Retries int
}

// Client for RNG service with failover support
type Client struct {
	ServiceURL string // Keep for backward compatibility

	// Failover support
	Services       []ServiceEndpoint
	CurrentIndex   int
	States         []ServiceState
	Mutex          sync.RWMutex
	HealthChecker  *HealthChecker
	TelegramClient interface{} // Will be set to *telegram.Client when available
}

// HealthChecker manages health monitoring for services
type HealthChecker struct {
	Services         []ServiceEndpoint
	CheckInterval    time.Duration
	SuccessThreshold int
	FailureThreshold int
	Client           *Client
}

// NewClient creates a new RNG client (backward compatible)
func NewClient(serviceURL string) *Client {
	return &Client{
		ServiceURL: serviceURL,
		Services: []ServiceEndpoint{
			{URL: serviceURL, Timeout: 10 * time.Second, Retries: 3},
		},
		CurrentIndex: 0,
		States:       []ServiceState{HEALTHY},
	}
}

// NewFailoverClient creates a new RNG client with failover support
func NewFailoverClient(cfg interface{}, telegramClient interface{}) *Client {
	// Type assertion to get config fields - using the actual config.Config type
	config := cfg.(config.Config)

	services := []ServiceEndpoint{
		{URL: config.RNGServiceURL, Timeout: 10 * time.Second, Retries: 3},
	}
	states := []ServiceState{HEALTHY}

	// Add secondary service if configured
	if config.RNG2ServiceURL != "" {
		services = append(services, ServiceEndpoint{
			URL:     config.RNG2ServiceURL,
			Timeout: 10 * time.Second,
			Retries: 3,
		})
		states = append(states, HEALTHY)
	}

	// Add tertiary service if configured
	if config.RNG3ServiceURL != "" {
		services = append(services, ServiceEndpoint{
			URL:     config.RNG3ServiceURL,
			Timeout: 10 * time.Second,
			Retries: 3,
		})
		states = append(states, HEALTHY)
	}

	client := &Client{
		ServiceURL:     config.RNGServiceURL, // Keep for backward compatibility
		Services:       services,
		CurrentIndex:   0,
		States:         states,
		TelegramClient: telegramClient,
	}

	// Create health checker
	client.HealthChecker = &HealthChecker{
		Services:         services,
		CheckInterval:    30 * time.Second,
		SuccessThreshold: 3,
		FailureThreshold: 2,
		Client:           client,
	}

	// Start health checker
	go client.startHealthChecker()

	return client
}

type Request struct {
	ClientID         string  `json:"client_id"`
	GameID           string  `json:"game_id"`
	PlayerID         string  `json:"player_id"`
	BetID            string  `json:"bet_id"`
	RTP              float64 `json:"rtp"`
	PayoutMultiplier float64 `json:"payout_multiplier"`
	RequestSalt      string  `json:"request_salt"`
	BetAmount        float64 `json:"bet_amount"`
	IPAddress        string  `json:"ip_address"`
	UserAgent        string  `json:"user_agent"`
	FeatureBuy       bool    `json:"feature_buy,omitempty"`
}

type Response struct {
	PrefOutcome string  `json:"pref_outcome"`
	WinAmount   float64 `json:"win_amount"`
	WinProb     float64 `json:"win_prob"`
}

// GetOutcome calls the RNG service and returns the outcome with failover support
func (c *Client) GetOutcome(clientID, gameID, playerID, betID string, rtp, payoutMultiplier, betAmount float64, ipAddress string, userAgent string, featureBuy bool) (Response, error) {
	c.Mutex.RLock()
	currentIndex := c.CurrentIndex
	c.Mutex.RUnlock()

	// Try current service
	resp, err := c.tryService(currentIndex, clientID, gameID, playerID, betID, rtp, payoutMultiplier, betAmount, ipAddress, userAgent, featureBuy)
	if err == nil {
		return resp, nil
	}

	// Current service failed, try failover
	return c.failoverToNext(currentIndex, err, clientID, gameID, playerID, betID, rtp, payoutMultiplier, betAmount, ipAddress, userAgent, featureBuy)
}

// tryService attempts to call a specific service
func (c *Client) tryService(serviceIndex int, clientID, gameID, playerID, betID string, rtp, payoutMultiplier, betAmount float64, ipAddress string, userAgent string, featureBuy bool) (Response, error) {
	if serviceIndex >= len(c.Services) {
		return Response{}, fmt.Errorf("invalid service index: %d", serviceIndex)
	}

	service := c.Services[serviceIndex]

	reqBody, err := json.Marshal(Request{
		ClientID:         clientID,
		GameID:           gameID,
		BetID:            betID,
		PlayerID:         playerID,
		RTP:              rtp,
		PayoutMultiplier: payoutMultiplier,
		RequestSalt:      uuid.New().String(),
		BetAmount:        betAmount,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		FeatureBuy:       featureBuy,
	})
	if err != nil {
		log.Printf("Error marshaling RNG request for service %d: %v", serviceIndex, err)
		return Response{}, err
	}

	log.Printf("RNG request to service %d (%s): %s", serviceIndex, service.URL, string(reqBody))

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: service.Timeout,
	}

	resp, err := httpClient.Post(service.URL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error calling RNG API service %d (%s): %v", serviceIndex, service.URL, err)
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("RNG API service %d (%s) returned non-200 status: %d", serviceIndex, service.URL, resp.StatusCode)
		return Response{}, errors.New("RNG API call failed")
	}

	var rngResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rngResp); err != nil {
		log.Printf("Error decoding RNG response from service %d: %v", serviceIndex, err)
		return Response{}, err
	}

	log.Printf("RNG response from service %d: %+v", serviceIndex, rngResp)
	return rngResp, nil
}

// failoverToNext handles failover to the next available service
func (c *Client) failoverToNext(currentIndex int, lastErr error, clientID, gameID, playerID, betID string, rtp, payoutMultiplier, betAmount float64, ipAddress string, userAgent string, featureBuy bool) (Response, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Mark current service as failed
	if currentIndex < len(c.States) {
		c.States[currentIndex] = FAILED
		log.Printf("Marked RNG service %d as FAILED due to error: %v", currentIndex, lastErr)
	}

	// Try next available service
	for i := 1; i < len(c.Services); i++ {
		nextIndex := (currentIndex + i) % len(c.Services)

		if c.States[nextIndex] == HEALTHY || c.States[nextIndex] == DEGRADED {
			// Switch to next service
			c.CurrentIndex = nextIndex

			// Send Telegram notification
			c.sendFailoverNotification(currentIndex, nextIndex, lastErr)

			log.Printf("Switching RNG from service %d to service %d", currentIndex, nextIndex)

			// Try the new service
			return c.tryService(nextIndex, clientID, gameID, playerID, betID, rtp, payoutMultiplier, betAmount, ipAddress, userAgent, featureBuy)
		}
	}

	// All services failed
	c.sendAllServicesFailedNotification(lastErr)
	return Response{}, fmt.Errorf("all RNG services failed, last error: %v", lastErr)
}

// startHealthChecker starts the background health monitoring
func (c *Client) startHealthChecker() {
	if c.HealthChecker == nil {
		return
	}

	ticker := time.NewTicker(c.HealthChecker.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.checkAllServices()
	}
}

// checkAllServices checks the health of all services
func (c *Client) checkAllServices() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	for i, service := range c.Services {
		if c.States[i] == FAILED {
			// Test failed service
			if c.testService(service) {
				c.States[i] = RECOVERING
				log.Printf("RNG service %d recovered, testing...", i)

				// Test multiple times before marking as healthy
				if c.testServiceMultipleTimes(service, c.HealthChecker.SuccessThreshold) {
					c.States[i] = HEALTHY
					c.sendServiceRecoveredNotification(i)
					log.Printf("RNG service %d is now healthy", i)
				}
			}
		}
	}
}

// testService tests if a service is responding
func (c *Client) testService(service ServiceEndpoint) bool {
	// Create a simple test request
	testReq := Request{
		ClientID:         "279",
		GameID:           "46",
		PlayerID:         "284",
		BetID:            "bet111",
		RTP:              0.9,
		PayoutMultiplier: 0.2,
		RequestSalt:      uuid.New().String(),
		BetAmount:        100,
		IPAddress:        "127.0.0.1",
		UserAgent:        "health_check",
		FeatureBuy:       false,
	}

	reqBody, err := json.Marshal(testReq)
	if err != nil {
		return false
	}

	httpClient := &http.Client{
		Timeout: 5 * time.Second, // Shorter timeout for health checks
	}

	resp, err := httpClient.Post(service.URL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// testServiceMultipleTimes tests a service multiple times to ensure it's stable
func (c *Client) testServiceMultipleTimes(service ServiceEndpoint, times int) bool {
	successCount := 0
	for i := 0; i < times; i++ {
		if c.testService(service) {
			successCount++
		}
		time.Sleep(1 * time.Second) // Small delay between tests
	}
	return successCount >= times
}

// sendFailoverNotification sends a Telegram notification about service failover
func (c *Client) sendFailoverNotification(fromIndex, toIndex int, err error) {
	if c.TelegramClient == nil {
		return
	}

	// Use reflection to call SendMessage method
	if sendMsg, ok := c.TelegramClient.(interface{ SendMessage(string) error }); ok {
		message := fmt.Sprintf(`
🚨 RNG Service Failover Alert

Switched: S%d → S%d
Reason: %v
Time: %s
Status: Using backup service
		`, fromIndex+1, toIndex+1, err, time.Now().Format("2006-01-02 15:04:05 UTC"))

		if telegramErr := sendMsg.SendMessage(message); telegramErr != nil {
			log.Printf("Failed to send Telegram notification: %v", telegramErr)
		}
	}
}

// sendServiceRecoveredNotification sends a Telegram notification about service recovery
func (c *Client) sendServiceRecoveredNotification(serviceIndex int) {
	if c.TelegramClient == nil {
		return
	}

	// Use reflection to call SendMessage method
	if sendMsg, ok := c.TelegramClient.(interface{ SendMessage(string) error }); ok {
		message := fmt.Sprintf(`
✅ RNG Service Recovery Alert

Recovered: S%d
Time: %s
Status: Service is healthy again
		`, serviceIndex+1, time.Now().Format("2006-01-02 15:04:05 UTC"))

		if telegramErr := sendMsg.SendMessage(message); telegramErr != nil {
			log.Printf("Failed to send Telegram notification: %v", telegramErr)
		}
	}
}

// sendAllServicesFailedNotification sends a Telegram notification when all services fail
func (c *Client) sendAllServicesFailedNotification(err error) {
	if c.TelegramClient == nil {
		return
	}

	// Use reflection to call SendMessage method
	if sendMsg, ok := c.TelegramClient.(interface{ SendMessage(string) error }); ok {
		message := fmt.Sprintf(`
🚨 CRITICAL: All RNG Services Failed

Last Error: %v
Time: %s
Status: All services are down - immediate attention required!
		`, err, time.Now().Format("2006-01-02 15:04:05 UTC"))

		if telegramErr := sendMsg.SendMessage(message); telegramErr != nil {
			log.Printf("Failed to send Telegram notification: %v", telegramErr)
		}
	}
}
