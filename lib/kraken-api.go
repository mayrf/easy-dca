package lib

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Order represents a single order entry [price, volume, timestamp]
type Order struct {
	Price     float64
	Volume    float64
	Timestamp float64
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

// OrderBook represents the asks and bids for a trading pair
type OrderBook struct {
	Asks []Order `json:"asks"`
	Bids []Order `json:"bids"`
}

// KrakenResponse represents the complete API response structure
type OrderBookResponse struct {
	Error  []string             `json:"error"`
	Result map[string]OrderBook `json:"result"`
}

// Helper methods for Order - now direct field access
func (o Order) GetPrice() float64     { return o.Price }
func (o Order) GetVolume() float64    { return o.Volume }
func (o Order) GetTimestamp() float64 { return o.Timestamp }

func trimFloat32ToOneDecimal(f float32) float32 {
	return float32(math.Round(float64(f)*10) / 10)
}


func AddOrder(pair string, price float32, volume float32, publicKey string, privateKey string, validate bool) {
	resp, err := request(&Request{
		Method: "POST",
		Path:   "/0/private/AddOrder",
		Body: map[string]any{
			"ordertype": "limit",
			"type":      "buy",
			"volume":    volume,
			// "pair":        "BTC/USD",
			"pair":   pair,
			"price":  trimFloat32ToOneDecimal(price),
			"oflags": "post",
			// "timeinforce": "GTD",
			// "expiretm":    "+5",
			"validate": validate,
		},
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Environment: "https://api.kraken.com",
	})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", data)

}

// func GetOrderBook(pair string, count int) map[string]interface{} []byte {
func GetOrderBook(pair string, count int) OrderBookResponse {
	resp, err := request(&Request{
		Method: "GET",
		Path:   "/0/public/Depth",
		Query: map[string]any{
			"pair":  pair,
			"count": count,
		},
		Environment: "https://api.kraken.com",
	})
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	jsonData, err := io.ReadAll(resp.Body)

	var response OrderBookResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Fatal(err)
	}

	// Check for API errors
	if len(response.Error) > 0 {
		fmt.Printf("API Error: %v\n", response.Error)
		// return
	}

	return response

}

type Request struct {
	Method      string
	Path        string
	Query       map[string]any
	Body        map[string]any
	PublicKey   string
	PrivateKey  string
	Environment string
}

func request(c *Request) (*http.Response, error) {
	url := c.Environment + c.Path
	var queryString string
	if len(c.Query) > 0 {
		queryValues, err := mapToURLValues(c.Query)
		if err != nil {
			return nil, fmt.Errorf("query to URL values: %s", err)
		}
		queryString = queryValues.Encode()
		url += "?" + queryString
	}
	var nonce any
	bodyMap := c.Body
	if len(c.PublicKey) > 0 {
		if bodyMap == nil {
			bodyMap = make(map[string]any)
		}
		var ok bool
		nonce, ok = bodyMap["nonce"]
		if !ok {
			nonce = getNonce()
			bodyMap["nonce"] = nonce
		}
	}
	headers := make(http.Header)
	var bodyReader io.Reader
	var bodyString string
	if len(bodyMap) > 0 {
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("json marshal: %s", err)
		}
		bodyString = string(bodyBytes)
		bodyReader = bytes.NewReader(bodyBytes)
		headers.Set("Content-Type", "application/json")
	}
	request, err := http.NewRequest(c.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("http new request: %s", err)
	}
	if len(c.PublicKey) > 0 {
		signature, err := getSignature(c.PrivateKey, queryString+bodyString, fmt.Sprint(nonce), c.Path)
		if err != nil {
			return nil, fmt.Errorf("get signature: %s", err)
		}
		headers.Set("API-Key", c.PublicKey)
		headers.Set("API-Sign", signature)
	}
	request.Header = headers
	return http.DefaultClient.Do(request)
}

func getNonce() string {
	return fmt.Sprint(time.Now().UnixMilli())
}

func getSignature(privateKey string, data string, nonce string, path string) (string, error) {
	message := sha256.New()
	message.Write([]byte(nonce + data))
	return sign(privateKey, []byte(path+string(message.Sum(nil))))
}

func sign(privateKey string, message []byte) (string, error) {
	key, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", err
	}
	hmacHash := hmac.New(sha512.New, key)
	hmacHash.Write(message)
	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil)), nil
}

func mapToURLValues(m map[string]any) (url.Values, error) {
	uv := make(url.Values)
	for k, v := range m {
		switch v := v.(type) {
		case []string:
			uv[k] = v
		case string:
			uv[k] = []string{v}
		default:
			j, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			uv[k] = []string{string(j)}
		}
	}
	return uv, nil
}
