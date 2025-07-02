#!/usr/bin/env bash

# Example: Run easy-dca once with ntfy notification
# Set your own values for the variables below
export EASY_DCA_PUBLIC_KEY="your_public_key"
export EASY_DCA_PRIVATE_KEY="your_private_key"
export NOTIFY_METHOD="ntfy"
export NOTIFY_NTFY_TOPIC="yourtopic"
# export NOTIFY_NTFY_URL="https://ntfy.sh" # Optional, defaults to ntfy.sh

# Optionally set trading config
# export EASY_DCA_PRICEFACTOR="0.998"
# export EASY_DCA_MONTHLY_VOLUME="140.0"

# Run the app once
exec go run ./cmd/easy-dca 