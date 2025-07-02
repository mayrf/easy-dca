#!/usr/bin/env bash

# Example: Run easy-dca every minute
# Make sure you have a .env file with your configuration
# See .env.example for required variables
# Set EASY_DCA_CRON="* * * * *" in your .env file for every minute

# Run the app in cron mode
exec go run ./cmd/easy-dca 