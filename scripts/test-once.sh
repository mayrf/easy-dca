#!/usr/bin/env bash

# Example: Run easy-dca once
# Make sure you have a .env file with your configuration
# See .env.example for required variables
export EASY_DCA_SCHEDULER_MODE=manual
# Run the app once
exec go run ./cmd/easy-dca 