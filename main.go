package main

import (
	"log"
	"os"
	"strconv"
	"github.com/joho/godotenv"
	"github.com/mayrf/easy-dca/lib"
)

type OrderBook struct {
	Asks [][]float64 `json:"asks"`
	Bids [][]float64 `json:"bids"`
}

type OrderBookResponse struct {
	Eror [][]float64 `json:"asks"`
	Bids [][]float64 `json:"bids"`
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func main() {
	log.Print("Trying to load environment variables from '.env' file")
	err := godotenv.Load()
	if err != nil {
		log.Print("Not found .env file")
	}
	pair := "BTC/EUR"
	publicKey := os.Getenv("EASY_DCA_PUBLIC_KEY")
	privateKey := os.Getenv("EASY_DCA_PRIVATE_KEY")
	order_validation := getEnvAsBool("EASY_DCA_VALIDATION_ON", true)
	priceFactor := getEnvAsFloat32("EASY_DCA_PRICEFACTOR", 0.998)
	monthlyVolume := getEnvAsFloat32("EASY_DCA_MONTHLY_VOLUME", 140.0)
	dailyVolume := monthlyVolume / 30.0

	if priceFactor >= 1 {
		panic("priceFactor must be smaller than 1 in order to place a limit order as a maker")
	}

	log.Printf("Fetching orders for %s", pair)
	response := lib.GetOrderBook(pair, 10)
	orderBook := response.Result[pair]
	log.Printf("Best Ask: Price=%.2f, Volume=%.3f\n",
		orderBook.Asks[0].Price, orderBook.Asks[0].Volume)

	log.Printf("Best Bid: Price=%.2f, Volume=%.3f\n",
		orderBook.Bids[0].Price, orderBook.Bids[0].Volume)

	buyPrice := priceFactor * float32(orderBook.Asks[0].Price)
	buyVolume := dailyVolume / buyPrice
	if buyVolume < 0.00005 {
		log.Printf("Order volume of %.8f BTC is too small. Miminum is 0.00005 BTC", buyVolume)
		log.Printf("Setting order volume to 0.00005 BTC")
		buyVolume = 0.00005

	}
	log.Printf("Ordering price factor: %.4f, Ordering Price: %.2f", priceFactor, buyPrice)
	log.Printf("Ordering %.8f BTC at a price of %.2f for a total of %.2f Euro", buyVolume, buyPrice, buyPrice*buyVolume)
	lib.AddOrder("BTC/EUR", float32(buyPrice), float32(buyVolume), publicKey, privateKey, order_validation)
}
