#!/usr/bin/env bash

# Example: Run easy-dca every 5 minutes using the --cron flag
# Make sure you have a .env file with your configuration
# See .env.example for required variables

# Run the app in cron mode every 5 minutes using the CLI flag
exec go run ./cmd/easy-dca --cron "*/5 * * * *" 