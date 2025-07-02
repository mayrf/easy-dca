// Command easy-dca is the entrypoint for the DCA trading application using the Kraken API.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/mayrf/easy-dca/internal/kraken"
	"github.com/mayrf/easy-dca/internal/order"
	"github.com/mayrf/easy-dca/internal/config"
)

var Version = "dev"

// Notifier is an interface for sending notifications.
type Notifier interface {
	Notify(ctx context.Context, subject, message string) error
}

// NtfyNotifier sends notifications via ntfy.sh or a custom ntfy server.
type NtfyNotifier struct {
	Topic string
	URL   string
}

func (n *NtfyNotifier) Notify(ctx context.Context, subject, message string) error {
	if n.Topic == "" {
		return fmt.Errorf("ntfy topic is not set")
	}
	url := n.URL
	if url == "" {
		url = "https://ntfy.sh"
	}
	ntfyURL := fmt.Sprintf("%s/%s", strings.TrimRight(url, "/"), n.Topic)
	req, err := http.NewRequestWithContext(ctx, "POST", ntfyURL, strings.NewReader(message))
	if err != nil {
		return err
	}
	if subject != "" {
		req.Header.Set("Title", subject)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy notification failed: %s", resp.Status)
	}
	return nil
}

// getNotifier returns a Notifier based on config.
func getNotifier(cfg config.Config) Notifier {
	switch strings.ToLower(cfg.NotifyMethod) {
	case "ntfy":
		return &NtfyNotifier{Topic: cfg.NotifyNtfyTopic, URL: cfg.NotifyNtfyURL}
	// Add more cases for other notification methods (slack, email, etc.)
	default:
		return nil
	}
}

// runDCA performs one DCA cycle and sends a notification if configured.
func runDCA(cfg config.Config, notifier Notifier) {
	log.Printf("Fetching orders for %s", cfg.Pair)
	response, err := kraken.GetOrderBook(cfg.Pair, 10)
	if err != nil {
		log.Printf("Failed to fetch order book: %v", err)
		if notifier != nil {
			notifier.Notify(context.Background(), "DCA Error", fmt.Sprintf("Failed to fetch order book: %v", err))
		}
		return
	}
	orderBook := response.Result[cfg.Pair]
	log.Printf("Best Ask: Price=%.2f, Volume=%.3f\n",
		orderBook.Asks[0].Price, orderBook.Asks[0].Volume)

	log.Printf("Best Bid: Price=%.2f, Volume=%.3f\n",
		orderBook.Bids[0].Price, orderBook.Bids[0].Volume)

	buyPrice := cfg.PriceFactor * float32(orderBook.Asks[0].Price)
	buyVolume := cfg.DailyVolume / buyPrice
	if buyVolume < 0.00005 {
		log.Printf("Order volume of %.8f BTC is too small. Miminum is 0.00005 BTC", buyVolume)
		log.Printf("Setting order volume to 0.00005 BTC")
		buyVolume = 0.00005
	}
	log.Printf("Ordering price factor: %.4f, Ordering Price: %.2f", cfg.PriceFactor, buyPrice)
	log.Printf("Ordering %.8f BTC at a price of %.2f for a total of %.2f Euro", buyVolume, buyPrice, buyPrice*buyVolume)
	if cfg.DryRun {
		log.Printf("Dry run mode: order will only be validated, not executed.")
	}
	if err := kraken.AddOrder(cfg.Pair, float32(buyPrice), float32(buyVolume), cfg.PublicKey, cfg.PrivateKey, !cfg.DryRun); err != nil {
		log.Printf("Failed to add order: %v", err)
		if notifier != nil {
			notifier.Notify(context.Background(), "DCA Error", fmt.Sprintf("Failed to add order: %v", err))
		}
		return
	}
	msg := fmt.Sprintf("Ordered %.8f BTC at %.2f EUR (total %.2f EUR)", buyVolume, buyPrice, buyPrice*buyVolume)
	if notifier != nil {
		err := notifier.Notify(context.Background(), "DCA Success", msg)
		if err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}
}

// main is the entrypoint for the easy-dca CLI application.
func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println("easy-dca version:", Version)
		return
	}

	log.Print("Trying to load environment variables from '.env' file")
	err := godotenv.Load()
	if err != nil {
		log.Print("Not found .env file")
	}

	cronFlag := flag.String("cron", "", "Cron expression for scheduling (overrides EASY_DCA_CRON)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// CLI flag overrides env
	cronExpr := *cronFlag
	if cronExpr == "" {
		cronExpr = cfg.CronExpr
	}

	notifier := getNotifier(cfg)

	if cronExpr != "" {
		c := cron.New()
		_, err := c.AddFunc(cronExpr, func() { runDCA(cfg, notifier) })
		if err != nil {
			log.Fatalf("Invalid cron expression: %v", err)
		}
		log.Printf("Running in cron mode: %s", cronExpr)
		c.Start()
		select {} // Block forever
	} else {
		runDCA(cfg, notifier)
	}
}
