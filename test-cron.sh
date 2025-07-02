#!/usr/bin/env bash

# Example: Run easy-dca every minute with ntfy notification
# Set your own values for the variables below
export EASY_DCA_PUBLIC_KEY="your_public_key"
export EASY_DCA_PRIVATE_KEY="your_private_key"
export NOTIFY_METHOD="ntfy"
export NOTIFY_NTFY_TOPIC="yourtopic"
# export NOTIFY_NTFY_URL="https://ntfy.sh" # Optional, defaults to ntfy.sh

# Optionally set trading config
# export EASY_DCA_PRICEFACTOR="0.998"
# export EASY_DCA_MONTHLY_VOLUME="140.0"

# Set cron expression for every minute
export EASY_DCA_CRON="* * * * *"

# Run the app in cron mode
exec go run ./cmd/easy-dca 