package order

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestOrderUnmarshalJSON(t *testing.T) {
	jsonData := `["12345.67", "0.01", 1680000000]`
	var o Order
	if err := json.Unmarshal([]byte(jsonData), &o); err != nil {
		t.Fatalf("failed to unmarshal Order: %v", err)
	}
	if o.Price != 12345.67 {
		t.Errorf("expected Price 12345.67, got %v", o.Price)
	}
	if o.Volume != 0.01 {
		t.Errorf("expected Volume 0.01, got %v", o.Volume)
	}
	if o.Timestamp != 1680000000 {
		t.Errorf("expected Timestamp 1680000000, got %v", o.Timestamp)
	}
}

func TestOrderHelperMethods(t *testing.T) {
	o := Order{Price: 100.5, Volume: 0.5, Timestamp: 1234567890}
	if o.GetPrice() != 100.5 {
		t.Errorf("GetPrice() = %v, want 100.5", o.GetPrice())
	}
	if o.GetVolume() != 0.5 {
		t.Errorf("GetVolume() = %v, want 0.5", o.GetVolume())
	}
	if o.GetTimestamp() != 1234567890 {
		t.Errorf("GetTimestamp() = %v, want 1234567890", o.GetTimestamp())
	}
}

func TestOrderBookStruct(t *testing.T) {
	ob := OrderBook{
		Asks: []Order{{Price: 1, Volume: 2, Timestamp: 3}},
		Bids: []Order{{Price: 4, Volume: 5, Timestamp: 6}},
	}
	if len(ob.Asks) != 1 || len(ob.Bids) != 1 {
		t.Errorf("unexpected asks or bids length: %+v", ob)
	}
	if !reflect.DeepEqual(ob.Asks[0], Order{Price: 1, Volume: 2, Timestamp: 3}) {
		t.Errorf("unexpected ask: %+v", ob.Asks[0])
	}
	if !reflect.DeepEqual(ob.Bids[0], Order{Price: 4, Volume: 5, Timestamp: 6}) {
		t.Errorf("unexpected bid: %+v", ob.Bids[0])
	}
} 
