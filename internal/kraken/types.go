// Package kraken provides types for Kraken API responses.
package kraken

import "github.com/mayrf/easy-dca/internal/order"

// OrderBookResponse represents the complete API response structure from Kraken
// It uses the generic OrderBook type from the order package
type OrderBookResponse struct {
	Error  []string                `json:"error"` // List of error messages from the API
	Result map[string]order.OrderBook `json:"result"` // Map of trading pair to order book
} 