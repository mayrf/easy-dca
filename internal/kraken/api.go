// Package kraken provides integration with the Kraken cryptocurrency exchange API.
//
// ⚠️  EXPERIMENTAL API INTEGRATION: This Kraken API integration is experimental
// and may contain bugs or errors. Order placement logic may fail, price calculations
// may be incorrect, and API calls may timeout or return unexpected results.
// The maintainers take NO responsibility for failed trades, incorrect orders,
// financial losses, or any damages resulting from the use of this API integration.
// USE AT YOUR OWN RISK.
//
// See LICENSE file for full legal terms and conditions.
package kraken

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

// Request represents an HTTP request to the Kraken API.
type Request struct {
	Method      string
	Path        string
	Query       map[string]any
	Body        map[string]any
	PublicKey   string
	PrivateKey  string
	Environment string
}

// AddOrder places a new limit buy order on Kraken.
// Returns the parsed response and an error if the request fails or the API returns an error.
func AddOrder(pair string, price float32, volume float32, publicKey string, privateKey string, validate bool) (*AddOrderResponse, error) {
	resp, err := request(&Request{
		Method: "POST",
		Path:   "/0/private/AddOrder",
		Body: map[string]any{
			"ordertype": "limit",
			"type":      "buy",
			"volume":    volume,
			"pair":   pair,
			"price":  trimFloat32ToOneDecimal(price),
			"oflags": "post",
			"validate": validate,
		},
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Environment: "https://api.kraken.com",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var response AddOrderResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response.Error) > 0 {
		return &response, fmt.Errorf("API Error: %v", response.Error)
	}
	
	return &response, nil
}

// GetOrderBook fetches the order book for a trading pair from Kraken.
// Returns the order book response or an error.
func GetOrderBook(pair string, count int) (OrderBookResponse, error) {
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
		return OrderBookResponse{}, err
	}

	defer resp.Body.Close()
	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		return OrderBookResponse{}, err
	}

	var response OrderBookResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		return OrderBookResponse{}, err
	}

	if len(response.Error) > 0 {
		return response, fmt.Errorf("API Error: %v", response.Error)
	}

	return response, nil
}

// trimFloat32ToOneDecimal rounds down a float32 to one decimal place.
func trimFloat32ToOneDecimal(f float32) float32 {
	return float32(math.Floor(float64(f)*10) / 10)
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

// FormatOrderResponse creates a nice log message from the AddOrder response
func FormatOrderResponse(response *AddOrderResponse, isDryRun bool) string {
	if response == nil {
		return "No response received"
	}
	
	mode := "DRY RUN"
	if !isDryRun {
		mode = "LIVE ORDER"
	}
	
	orderDesc := response.Result.Descr.Order
	
	if len(response.Result.Txid) > 0 {
		return fmt.Sprintf("[%s] Order placed successfully! Transaction ID: %s | %s", 
			mode, response.Result.Txid[0], orderDesc)
	} else {
		return fmt.Sprintf("[%s] Order validated successfully! | %s", 
			mode, orderDesc)
	}
} 