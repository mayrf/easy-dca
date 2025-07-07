// Command easy-dca is the entrypoint for the DCA trading application using the Kraken API.
//
// ⚠️  DISCLAIMER: This software is provided "AS IS" without any warranties.
// Trading cryptocurrencies involves substantial risk of loss. You can lose some or all
// of your invested capital. The maintainers take NO responsibility for financial losses,
// API failures, software bugs, or any other damages. USE AT YOUR OWN RISK.
// ONLY TRADE WITH MONEY YOU CAN AFFORD TO LOSE.
//
// See LICENSE file for full legal terms and conditions.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mayrf/easy-dca/internal/config"
	"github.com/mayrf/easy-dca/internal/dca"
	"github.com/mayrf/easy-dca/internal/notifications"
	"github.com/mayrf/easy-dca/internal/scheduler"
)

var Version = "dev"

// main is the entrypoint for the easy-dca CLI application.
func main() {
	config.ConfigureLogging()

	versionFlag := flag.Bool("version", false, "Print version and exit")
	cronFlag := flag.String("cron", "", "Cron expression for scheduling (overrides EASY_DCA_CRON)")
	flag.Parse()
	
	if *versionFlag {
		fmt.Println("easy-dca version:", Version)
		return
	}

	err := godotenv.Load()
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		log.Printf("Error loading .env file: %v (continuing with process environment)", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// CLI flag overrides env
	if *cronFlag != "" {
		cfg.CronExpr = *cronFlag
		log.Printf("⏰ Cron expression overridden by CLI flag: %s", *cronFlag)
	}

	// Create notifier
	notifier := notifications.CreateNotifier(cfg)

	// Create DCA runner
	runner := dca.NewRunner(cfg, notifier)

	// Create scheduler
	sched, err := scheduler.CreateScheduler(runner, cfg)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Print("Received shutdown signal, stopping scheduler...")
		cancel()
	}()

	// Start the scheduler
	log.Printf("Starting easy-dca with configuration: %s", cfg.Pair.String())
	if err := sched.Start(ctx); err != nil {
		log.Printf("Scheduler error: %v", err)
		os.Exit(1)
	}

	log.Print("easy-dca stopped")
}
