// Package order provides types and helpers for representing and working with orders and order books.
package order

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Order represents a single order entry [price, volume, timestamp].
type Order struct {
	Price     float64 // Order price
	Volume    float64 // Order volume
	Timestamp float64 // Order timestamp
}

// UnmarshalJSON implements json.Unmarshaler for Order
func (o *Order) UnmarshalJSON(data []byte) error {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 3 {
		return fmt.Errorf("expected array of 3 elements, got %d", len(arr))
	}

	// Convert each element to float64, handling both string and number types
	for i, val := range arr {
		var floatVal float64
		var err error

		switch v := val.(type) {
		case string:
			floatVal, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("failed to parse string to float: %v", err)
			}
		case float64:
			floatVal = v
		case int:
			floatVal = float64(v)
		default:
			return fmt.Errorf("unexpected type %T for array element", v)
		}

		// Assign to appropriate field
		switch i {
		case 0:
			o.Price = floatVal
		case 1:
			o.Volume = floatVal
		case 2:
			o.Timestamp = floatVal
		}
	}

	return nil
}

// OrderBook represents the asks and bids for a trading pair.
type OrderBook struct {
	Asks []Order `json:"asks"` // List of ask orders
	Bids []Order `json:"bids"` // List of bid orders
}

// GetPrice returns the price of the order.
func (o Order) GetPrice() float64     { return o.Price }
// GetVolume returns the volume of the order.
func (o Order) GetVolume() float64    { return o.Volume }
// GetTimestamp returns the timestamp of the order.
func (o Order) GetTimestamp() float64 { return o.Timestamp } 