// Package kraken provides types for Kraken API responses.
package kraken

import "github.com/mayrf/easy-dca/internal/order"

// OrderBookResponse represents the complete API response structure from Kraken
// It uses the generic OrderBook type from the order package
type OrderBookResponse struct {
	Error  []string                `json:"error"` // List of error messages from the API
	Result map[string]order.OrderBook `json:"result"` // Map of trading pair to order book
}

// AddOrderResponse represents the response from the AddOrder API call
type AddOrderResponse struct {
	Error  []string `json:"error"`  // List of error messages from the API
	Result struct {
		Txid  []string `json:"txid"`  // Transaction IDs (empty for dry run)
		Descr struct {
			Order string `json:"order"` // Order description
		} `json:"descr"`
	} `json:"result"`
} 