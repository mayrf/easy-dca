package kraken

import (
	"fmt"
	"testing"
	"time"

	"github.com/mayrf/easy-dca/internal/order"
	"github.com/mayrf/easy-dca/internal/config"
)

// TestGetOrderBookIntegration tests the GetOrderBook function with all supported trading pairs
// This is an integration test that makes real API calls to Kraken's public API
func TestGetOrderBookIntegration(t *testing.T) {
	// Use supported trading pairs from config
	supportedPairs := config.GetSupportedPairs()

	for _, pair := range supportedPairs {
		t.Run(pair, func(t *testing.T) {
			testGetOrderBookForPair(t, pair)
		})
	}
}

// testGetOrderBookForPair tests the GetOrderBook function for a specific trading pair
func testGetOrderBookForPair(t *testing.T, pair string) {
	// Test with different count values
	counts := []int{1, 5, 10, 25}

	for _, count := range counts {
		t.Run(fmt.Sprintf("count_%d", count), func(t *testing.T) {
			// Add a small delay to avoid rate limiting
			time.Sleep(100 * time.Millisecond)

			response, err := GetOrderBook(pair, count)
			if err != nil {
				t.Fatalf("GetOrderBook(%s, %d) failed: %v", pair, count, err)
			}

			// Validate response structure
			if len(response.Error) > 0 {
				t.Fatalf("API returned errors: %v", response.Error)
			}

			// Check that the result contains the expected trading pair
			orderBook, exists := response.Result[pair]
			if !exists {
				t.Fatalf("Response does not contain order book for pair %s", pair)
			}

			// Validate order book structure
			validateOrderBook(t, pair, orderBook, count)
		})
	}
}

// validateOrderBook validates the structure and content of an order book
func validateOrderBook(t *testing.T, pair string, orderBook order.OrderBook, expectedCount int) {
	// Check that we have both asks and bids
	if len(orderBook.Asks) == 0 {
		t.Errorf("Order book for %s has no asks", pair)
	}
	if len(orderBook.Bids) == 0 {
		t.Errorf("Order book for %s has no bids", pair)
	}

	// Check that we don't have more orders than requested (Kraken may return fewer)
	if len(orderBook.Asks) > expectedCount {
		t.Errorf("Order book for %s has %d asks, expected at most %d", pair, len(orderBook.Asks), expectedCount)
	}
	if len(orderBook.Bids) > expectedCount {
		t.Errorf("Order book for %s has %d bids, expected at most %d", pair, len(orderBook.Bids), expectedCount)
	}

	// Validate ask orders (should be sorted by price ascending)
	for i, ask := range orderBook.Asks {
		if ask.Price <= 0 {
			t.Errorf("Ask %d for %s has invalid price: %f", i, pair, ask.Price)
		}
		if ask.Volume <= 0 {
			t.Errorf("Ask %d for %s has invalid volume: %f", i, pair, ask.Volume)
		}
		if ask.Timestamp <= 0 {
			t.Errorf("Ask %d for %s has invalid timestamp: %f", i, pair, ask.Timestamp)
		}

		// Check that asks are sorted by price (ascending)
		if i > 0 && ask.Price < orderBook.Asks[i-1].Price {
			t.Errorf("Asks for %s are not sorted by price: %f < %f", pair, ask.Price, orderBook.Asks[i-1].Price)
		}
	}

	// Validate bid orders (should be sorted by price descending)
	for i, bid := range orderBook.Bids {
		if bid.Price <= 0 {
			t.Errorf("Bid %d for %s has invalid price: %f", i, pair, bid.Price)
		}
		if bid.Volume <= 0 {
			t.Errorf("Bid %d for %s has invalid volume: %f", i, pair, bid.Volume)
		}
		if bid.Timestamp <= 0 {
			t.Errorf("Bid %d for %s has invalid timestamp: %f", i, pair, bid.Timestamp)
		}

		// Check that bids are sorted by price (descending)
		if i > 0 && bid.Price > orderBook.Bids[i-1].Price {
			t.Errorf("Bids for %s are not sorted by price: %f > %f", pair, bid.Price, orderBook.Bids[i-1].Price)
		}
	}

	// Validate that best ask is higher than best bid (bid-ask spread)
	if len(orderBook.Asks) > 0 && len(orderBook.Bids) > 0 {
		bestAsk := orderBook.Asks[0].Price
		bestBid := orderBook.Bids[0].Price
		if bestAsk <= bestBid {
			t.Errorf("Invalid bid-ask spread for %s: best ask (%f) <= best bid (%f)", pair, bestAsk, bestBid)
		}
	}

	// Log some useful information for debugging
	t.Logf("Order book for %s: %d asks, %d bids", pair, len(orderBook.Asks), len(orderBook.Bids))
	if len(orderBook.Asks) > 0 && len(orderBook.Bids) > 0 {
		bestAsk := orderBook.Asks[0].Price
		bestBid := orderBook.Bids[0].Price
		spread := bestAsk - bestBid
		spreadPercent := (spread / bestAsk) * 100
		t.Logf("Best ask: %f, Best bid: %f, Spread: %f (%f%%)", bestAsk, bestBid, spread, spreadPercent)
	}
}

// TestGetOrderBookErrorHandling tests error handling for invalid trading pairs
func TestGetOrderBookErrorHandling(t *testing.T) {
	invalidPairs := []string{
		"INVALID/PAIR",
		"BTC/INVALID",
		// "ETH/EUR", // Not supported in our config, but valid on Kraken
		"",
		"BTC",
	}

	for _, pair := range invalidPairs {
		t.Run(pair, func(t *testing.T) {
			// Add a small delay to avoid rate limiting
			time.Sleep(100 * time.Millisecond)

			response, err := GetOrderBook(pair, 10)
			
			// For invalid pairs, we expect either an error or an API error response
			if err == nil {
				// If no error, check if the API returned an error response
				if len(response.Error) == 0 {
					t.Errorf("Expected error for invalid pair %s, but got successful response", pair)
				} else {
					t.Logf("API correctly returned error for invalid pair %s: %v", pair, response.Error)
				}
			} else {
				t.Logf("Correctly got error for invalid pair %s: %v", pair, err)
			}
		})
	}
} 