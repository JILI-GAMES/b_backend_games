package hilo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

var secretKey = []byte("hilo_secret_key_change_in_production")

// Card ranks with proper values (A=1, J=11, Q=12, K=13)
var ranks = []string{"ACE", "TWO", "THREE", "FOUR", "FIVE", "SIX", "SEVEN", "EIGHT", "NINE", "TEN", "JACK", "QUEEN", "KING"}
var suits = []string{"SPADES", "CLUBS", "DIAMONDS", "HEARTS"}

// GenerateDeck creates a shuffled deck based on seed
func GenerateDeck(seed string) []string {
	// Create deterministic seed
	h := sha256.Sum256([]byte(seed))
	seedInt := int64(0)
	for i := 0; i < 8; i++ {
		seedInt = (seedInt << 8) | int64(h[i])
	}
	
	rng := rand.New(rand.NewSource(seedInt))
	deck := []string{}
	
	for _, r := range ranks {
		for _, s := range suits {
			deck = append(deck, r+"_"+s)
		}
	}
	
	// Fisher-Yates shuffle
	for i := len(deck) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	
	return deck
}

// HashDeck creates a hash of the entire deck for verification
func HashDeck(deck []string) string {
	data := ""
	for _, c := range deck {
		data += c
	}
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// GetCardValue returns the numeric value of a card (A=1, J=11, Q=12, K=13)
func GetCardValue(card string) int {
	// Remove suit to get rank
	suit := card[len(card)-1:]
	rank := card[:len(card)-len(suit)]
	
	for i, r := range ranks {
		if r == rank {
			return i + 1 // A=1, 2=2, ..., K=13
		}
	}
	return -1
}

// ComputeHMAC computes HMAC for game state verification
func ComputeHMAC(seed string, position int, accumulated float64, betAmount float64) string {
	data := fmt.Sprintf("%s|%d|%.6f|%.6f", seed, position, accumulated, betAmount)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC verifies HMAC signature
func VerifyHMAC(seed string, position int, accumulated float64, betAmount float64, signature string) bool {
	expected := ComputeHMAC(seed, position, accumulated, betAmount)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// GenerateUniqueSeed generates a unique seed with timestamp and random component
func GenerateUniqueSeed() string {
	timestamp := time.Now().UnixNano()
	random := rand.Int63()
	return fmt.Sprintf("%d_%d", timestamp, random)
}