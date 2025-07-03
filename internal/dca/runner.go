// Package dca provides the core DCA trading logic.
package dca

import (
	"context"
	"fmt"
	"log"

	"github.com/mayrf/easy-dca/internal/config"
	"github.com/mayrf/easy-dca/internal/kraken"
	"github.com/mayrf/easy-dca/internal/notifications"
)

// Runner implements the DCARunner interface and contains the core DCA logic.
type Runner struct {
	cfg       config.Config
	notifier  notifications.Notifier
}

// NewRunner creates a new DCA runner with the given configuration and notifier.
func NewRunner(cfg config.Config, notifier notifications.Notifier) *Runner {
	return &Runner{
		cfg:      cfg,
		notifier: notifier,
	}
}

// RunDCA performs one DCA cycle and sends a notification if configured.
func (r *Runner) RunDCA() error {
	log.Printf("Fetching orders for %s", r.cfg.Pair)
	response, err := kraken.GetOrderBook(r.cfg.Pair, 10)
	if err != nil {
		log.Printf("Failed to fetch order book: %v", err)
		if r.notifier != nil {
			if err := r.notifier.Notify(context.Background(), "DCA Error", fmt.Sprintf("Failed to fetch order book: %v", err)); err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}
		return fmt.Errorf("failed to fetch order book: %w", err)
	}
	
	orderBook := response.Result[r.cfg.Pair]
	log.Printf("Best Ask: Price=%.2f, Volume=%.3f\n",
		orderBook.Asks[0].Price, orderBook.Asks[0].Volume)

	log.Printf("Best Bid: Price=%.2f, Volume=%.3f\n",
		orderBook.Bids[0].Price, orderBook.Bids[0].Volume)

	buyPrice := r.cfg.PriceFactor * float32(orderBook.Asks[0].Price)
	
	// Calculate fiat amount to spend based on configuration
	var fiatAmountToSpend float32
	if r.cfg.FiatAmountPerBuy > 0 {
		fiatAmountToSpend = r.cfg.FiatAmountPerBuy
	} else {
		fiatAmountToSpend = r.cfg.MonthlyFiatSpending / float32(r.cfg.BuysPerMonth)
	}
	
	btcQuantityToBuy := fiatAmountToSpend / buyPrice
	
	// Check if order size is close to minimum (within 10% of minimum)
	const (
		minBtcSize = 0.00005
		warningThreshold = 0.000055 // 10% above minimum
	)
	
	if btcQuantityToBuy < warningThreshold {
		log.Printf("Warning: Order size %.8f BTC is close to minimum (%.5f BTC)", btcQuantityToBuy, minBtcSize)
	}
	
	if btcQuantityToBuy < minBtcSize {
		if r.cfg.AutoAdjustMinOrder {
			log.Printf("Order volume of %.8f BTC is too small. Minimum is %.5f BTC", btcQuantityToBuy, minBtcSize)
			log.Printf("Auto-adjusting order volume to %.5f BTC", minBtcSize)
			btcQuantityToBuy = minBtcSize
			// Recalculate the actual fiat amount that will be spent
			actualFiatAmount := btcQuantityToBuy * buyPrice
			log.Printf("Note: This will actually spend %.2f EUR instead of the configured %.2f EUR", actualFiatAmount, fiatAmountToSpend)
		} else {
			log.Printf("Order volume of %.8f BTC is below minimum (%.5f BTC) and auto-adjustment is disabled", btcQuantityToBuy, minBtcSize)
			log.Printf("Order will likely fail, but cron job will continue running")
		}
	}
	
	log.Printf("Ordering price factor: %.4f, Ordering Price: %.2f", r.cfg.PriceFactor, buyPrice)
	log.Printf("Ordering %.8f BTC at a price of %.2f for a total of %.2f Euro", btcQuantityToBuy, buyPrice, btcQuantityToBuy*buyPrice)
	if r.cfg.DryRun {
		log.Printf("Dry run mode: order will only be validated, not executed.")
	}
	
	orderResponse, err := kraken.AddOrder(r.cfg.Pair, float32(buyPrice), float32(btcQuantityToBuy), r.cfg.PublicKey, r.cfg.PrivateKey, r.cfg.DryRun)
	if err != nil {
		log.Printf("Failed to add order: %v", err)
		if r.notifier != nil {
			if err := r.notifier.Notify(context.Background(), "DCA Error", fmt.Sprintf("Failed to add order: %v", err)); err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}
		return fmt.Errorf("failed to add order: %w", err)
	}
	
	// Log the formatted order response
	log.Print(kraken.FormatOrderResponse(orderResponse, r.cfg.DryRun))
	
	// Create notification message with order details
	var msg string
	if r.cfg.DryRun {
		msg = fmt.Sprintf("DRY RUN: Validated order for %.8f BTC at %.2f EUR (total %.2f EUR)", btcQuantityToBuy, buyPrice, fiatAmountToSpend)
	} else {
		if len(orderResponse.Result.Txid) > 0 {
			msg = fmt.Sprintf("LIVE ORDER: Placed order for %.8f BTC at %.2f EUR (total %.2f EUR) | TXID: %s", 
				btcQuantityToBuy, buyPrice, fiatAmountToSpend, orderResponse.Result.Txid[0])
		} else {
			msg = fmt.Sprintf("LIVE ORDER: Placed order for %.8f BTC at %.2f EUR (total %.2f EUR)", 
				btcQuantityToBuy, buyPrice, fiatAmountToSpend)
		}
	}
	
	if r.notifier != nil {
		err := r.notifier.Notify(context.Background(), "DCA Success", msg)
		if err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}
	
	return nil
} 