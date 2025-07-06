# easy-dca

A Go application for automated DCA (Dollar Cost Averaging) trading on Kraken, with cron scheduling and pluggable notifications (ntfy, etc).

## ⚠️ DISCLAIMER

**IMPORTANT: This software is provided "AS IS" without any warranties.**

### Financial Risk Disclaimer
- **Trading cryptocurrencies involves substantial risk of loss**
- **You can lose some or all of your invested capital**
- **Past performance does not guarantee future results**
- **The maintainers of this repository are NOT financial advisors**

### Software Disclaimer
- **This software may contain bugs or errors**
- **API integrations may fail or behave unexpectedly**
- **The maintainers take NO responsibility for:**
  - Financial losses from trading
  - API failures or incorrect orders
  - Software bugs or malfunctions
  - Data corruption or loss
  - Any other damages or losses

### Kraken API Disclaimer
- **Kraken API integration is experimental and may fail**
- **Order placement logic may contain errors**
- **Price calculations and order sizing may be incorrect**
- **API rate limits and errors are not fully handled**
- **The maintainers are NOT responsible for:**
  - Failed trades or missed opportunities
  - Incorrect order amounts or prices
  - API timeouts or connection issues
  - Account restrictions or suspensions

### Usage Agreement
By using this software, you acknowledge that:
- You understand the risks involved in cryptocurrency trading
- You are responsible for your own trading decisions
- You accept all risks and potential losses
- You will not hold the maintainers liable for any damages
- You have tested the software thoroughly before live trading

**USE AT YOUR OWN RISK. ONLY TRADE WITH MONEY YOU CAN AFFORD TO LOSE.**

### Legal Notice
This software is licensed under the GNU General Public License v3.0 (see LICENSE file). 
The GPL includes standard warranty disclaimers. The financial disclaimers above are 
in addition to those standard disclaimers and specifically address the risks of 
cryptocurrency trading and API integration.

## Features
- Run once or on a schedule (cron expression via CLI flag or env var)
- Pluggable notification system (ntfy supported, extensible for others)
- Simple configuration via environment variables
- Enhanced startup logging that explains your configuration in plain English
- Ready for Docker and docker-compose deployment (uses Chainguard images for security and minimal size)

## Quick Start (Docker Compose)

### 1. Clone the repository
```sh
git clone https://github.com/mayrf/easy-dca.git
cd easy-dca
```

### 2. Configure your environment
Copy the example environment file and customize it:
```sh
cp .env.example .env
```

### 3. Configure your API keys
For Docker secrets (recommended):
```sh
cp examples/public.key.example examples/public.key
cp examples/private.key.example examples/private.key
# Edit these files with your real API keys
```

### 3. Build and run with scripts
```sh
# Run once (manual mode)
./scripts/test-once.sh

# Run with cron flag
./scripts/test-cron-flag.sh
```

### 4. Build and run with docker compose
```sh
docker-compose up
```

**Note:** After the first build, if you make changes to the Go code, you'll need to rebuild:
```sh
docker-compose up --build
```

The app will run on the schedule you set in `EASY_DCA_CRON` and send notifications via ntfy.

**Tip:** When the application starts, it will display a comprehensive summary of your configuration, explaining what each setting does in plain English.

## Environment Variables
- `EASY_DCA_PUBLIC_KEY`: Kraken API public key (**required**)
- `EASY_DCA_PRIVATE_KEY`: Kraken API private key (**required**)

**Getting Kraken API Keys:**
- [How to Create a Kraken API Key](https://support.kraken.com/articles/360000919966-how-to-create-an-api-key)
- [Kraken API Documentation](https://docs.kraken.com/api/docs/rest-api/add-order)

**Required API Permissions:** `Orders and trades - Create & modify orders`

**Pro Tip:** If you have the [Kraken Pro app](https://www.kraken.com/features/cryptocurrency-apps) installed on your phone, you'll receive a notification once your DCA order has been filled!
- `EASY_DCA_PRICEFACTOR`: Price factor for limit orders (default: 0.998)
- `EASY_DCA_MONTHLY_FIAT_SPENDING`: Monthly fiat spending in EUR (optional, used if EASY_DCA_FIAT_AMOUNT_PER_BUY is not set)
- `EASY_DCA_FIAT_AMOUNT_PER_BUY`: Fixed fiat amount in EUR to spend each run (optional, takes precedence over EASY_DCA_MONTHLY_FIAT_SPENDING)
- `EASY_DCA_CRON`: Cron expression for scheduling (optional; if not set, runs once)
- `EASY_DCA_AUTO_ADJUST_MIN_ORDER`: If true, automatically adjust orders below minimum size (0.00005 BTC); if false, let them fail (default: false)
- `EASY_DCA_SCHEDULER_MODE`: Scheduler mode: "cron", "systemd", or "manual" (default: "cron" if EASY_DCA_CRON is set, otherwise "manual")
- `NOTIFY_METHOD`: Notification method (e.g., `ntfy`)
- `NOTIFY_NTFY_TOPIC`: ntfy topic (if using ntfy)
- `NOTIFY_NTFY_URL`: ntfy server URL (**required for ntfy notifications**; no default)
- `EASY_DCA_DRY_RUN`: If true (default), only validate orders (dry run); if false, actually place orders.
- `EASY_DCA_LOG_FORMAT`: Log format control (default: no timestamp)
  - `"timestamp"` or `"time"`: Standard format (2006/01/02 15:04:05)
  - `"microseconds"` or `"micro"`: Full datetime with microseconds (2006/01/02 15:04:05.000000)
  - Any other value or unset: No timestamp prefix

### Buy Amount Configuration

The app supports two ways to configure how much to buy:

1. **Fixed amount per buy**: Set `EASY_DCA_FIAT_AMOUNT_PER_BUY` to spend the same fiat amount each time
2. **Monthly amount**: Set `EASY_DCA_MONTHLY_FIAT_SPENDING` and the app will divide it by the number of buys per month (calculated from your cron schedule)

**Examples:**

- **Spend 10 EUR every day**: Set `EASY_DCA_FIAT_AMOUNT_PER_BUY=10` and `EASY_DCA_CRON="0 8 * * *"`
- **Spend 300 EUR per month, buying every 3 days**: Set `EASY_DCA_MONTHLY_FIAT_SPENDING=300` and `EASY_DCA_CRON="0 8 */3 * *"` (app will spend ~30 EUR each time)
- **Spend 150 EUR per month, buying weekly**: Set `EASY_DCA_MONTHLY_FIAT_SPENDING=150` and `EASY_DCA_CRON="0 8 * * 1"` (app will spend ~37.5 EUR each time)

**Note:** If both `EASY_DCA_FIAT_AMOUNT_PER_BUY` and `EASY_DCA_MONTHLY_FIAT_SPENDING` are set, the fixed amount per buy takes precedence.

### Minimum Order Size Behavior

Kraken has a minimum order size of 0.00005 BTC. The app handles this in two ways:

1. **Warning**: If your order size is below 0.000055 BTC (10% above minimum), you'll get a warning
2. **Auto-adjustment**: If your order is below 0.00005 BTC, the behavior depends on `EASY_DCA_AUTO_ADJUST_MIN_ORDER`:
   - **`false` (default)**: Order proceeds as-is and will likely fail, but the cron job continues running
   - **`true`**: Order size is automatically increased to 0.00005 BTC (you'll spend more fiat than configured)

**Example scenarios:**
- **Small daily DCA**: `EASY_DCA_FIAT_AMOUNT_PER_BUY=1` with daily cron and BTC at €50,000; 1 EUR per day = 0.00002 BTC → **Warning + likely failure**
- **Small weekly DCA**: `EASY_DCA_MONTHLY_FIAT_SPENDING=10` with weekly cron and BTC at €50,000. 2.5 EUR per week = 0.00005 BTC → **Warning + likely failure**
- **With auto-adjustment**: Same scenarios with `EASY_DCA_AUTO_ADJUST_MIN_ORDER=true` → **Order adjusted to 0.00005 BTC**

**Recommendation**: Use auto-adjustment only if you're comfortable spending more than your configured amount when BTC prices are high.

### Price Factor Strategy

The `EASY_DCA_PRICEFACTOR` determines at what percentage of the current best sell offer (ask price) your buy orders are placed:

- **Default: 0.998** (99.8% of ask price)
- **Range: 0.95 - 0.9999** (95% - 99.99% of ask price)

#### **How it works:**
- Current ask price: €50,000
- Price factor: 0.998
- Your buy order: €49,900 (99.8% of ask)

#### **Benefits:**
1. **Lower fees**: Limit orders below market price make you a "maker" (liquidity provider) with typically lower trading fees
2. **Better prices**: Bitcoin's volatility often creates opportunities to buy below current market prices
3. **Automated patience**: Orders wait for better prices rather than buying immediately

#### **Risks:**
1. **Unfilled orders**: If price never drops to your limit, orders may never execute
2. **Missed opportunities**: During strong uptrends, you might miss buying opportunities
3. **Timing risk**: Attempting to time market dips can backfire

#### **Recommended values:**
- **Conservative (0.995-0.9999)**: Higher fill probability, smaller savings
- **Balanced (0.99-0.995)**: Good balance of savings and fill probability  
- **Aggressive (0.95-0.99)**: Higher potential savings, lower fill probability

#### **Strategy considerations:**
- **Higher frequency DCA**: Use higher price factors (0.995+) for daily/weekly buys
- **Lower frequency DCA**: Can use lower price factors (0.95-0.99) for monthly buys
- **Market conditions**: Consider adjusting based on volatility and trend

## Secret Management: Using Key Path Variables

For better security, you can provide your Kraken API keys via file paths instead of directly in environment variables. This is especially useful when using Docker secrets, NixOS credentials, or other secret management systems.

- `EASY_DCA_PUBLIC_KEY_PATH`: Path to a file containing your Kraken API public key.
- `EASY_DCA_PRIVATE_KEY_PATH`: Path to a file containing your Kraken API private key.

If these variables are set, the app will read the key values from the files. This is preferred because:
- Secrets are not exposed in environment variables (which can be viewed by other processes or logged).
- Integrates well with Docker secrets, NixOS systemd credentials, and other secret managers.

### Example: Docker Compose with Secrets
```yaml
services:
  easy-dca:
    # ...
    secrets:
      - kraken-public-key
      - kraken-private-key
    environment:
      EASY_DCA_PUBLIC_KEY_PATH: "/run/secrets/kraken-public-key"
      EASY_DCA_PRIVATE_KEY_PATH: "/run/secrets/kraken-private-key"
      # ... other env vars ...
secrets:
  kraken-public-key:
    file: ./secrets/public.key
  kraken-private-key:
    file: ./secrets/private.key
```

### Example: NixOS with systemd credentials
```nix
services.easy-dca = {
  enable = true;
  credentials = {
    EASY_DCA_PUBLIC_KEY_PATH = "/run/secrets/kraken-public-key";
    EASY_DCA_PRIVATE_KEY_PATH = "/run/secrets/kraken-private-key";
  };
  environment = {
    # ... other env vars ...
  };
};
```

If both the `*_KEY_PATH` and the direct `*_KEY` variables are set, the path-based version takes precedence.

## CI/CD
- GitHub Actions workflow runs linting, tests, and builds the Docker image on every push and pull request to `master`.
- Linting is performed using `golangci-lint` to ensure code quality.

## Extending Notifications
To add more notification backends (Slack, Email, etc.), implement the `Notifier` interface in `cmd/easy-dca/easy-dca.go` and add a case to `getNotifier()`.

## NixOS Module Usage

If you use NixOS with flakes, you can enable and schedule `easy-dca` as a secure systemd timer service using the included NixOS module.

### 1. Add the Flake as an Input
In your system flake (e.g., `flake.nix`):

```nix
{
  inputs.easy-dca.url = "github:mayrf/easy-dca";
  # ... other inputs ...
}
```

### 2. Import the NixOS Module
In your `flake.nix` outputs:

```nix
{
  outputs = { self, nixpkgs, easy-dca, ... }@inputs: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        # ... other modules ...
        easy-dca.nixosModules.easy-dca
        ({ config, pkgs, ... }: {
          # Your configuration goes here
          services.easy-dca = {
            enable = true;
            schedule = "*-*-* 08:00:00";
            user = "easy-dca";
            group = "easy-dca";
            
            # Required API keys (using systemd credentials)
            publicKeyPath = "/run/secrets/kraken-public-key";
            privateKeyPath = "/run/secrets/kraken-private-key";
            
            # Trading configuration
            priceFactor = 0.998;
            dryRun = true; # Only validate orders (default)
            # dryRun = false; # Actually place orders
            autoAdjustMinOrder = false;
            
            # Buy amount configuration (choose one)
            fiatAmountPerBuy = 10.0; # Fixed amount per buy
            # monthlyFiatSpending = 300.0; # Monthly budget (alternative)
            
            # Notification configuration (optional)
            notifyMethod = "ntfy";
            notifyNtfyTopic = "yourtopic";
            notifyNtfyURL = "https://ntfy.sh";
            
            # Additional options
            # extraArgs = [ ];
          };
        })
      ];
    };
  };
}
```

**Note:** The `publicKeyPath` and `privateKeyPath` options should point to files containing your actual API keys. The module automatically loads these files as systemd credentials and makes them available to the application at runtime. You can use NixOS secrets management or place the key files in a secure location.

**Example with NixOS secrets:**
```nix
services.easy-dca = {
  enable = true;
  publicKeyPath = "/run/secrets/kraken-public-key";
  privateKeyPath = "/run/secrets/kraken-private-key";
  # ... other options
};

# Set up the secrets
sops.secrets = {
  "kraken-public-key" = {
    path = "/run/secrets/kraken-public-key";
  };
  "kraken-private-key" = {
    path = "/run/secrets/kraken-private-key";
  };
};
```

### 3. Rebuild and Switch
```sh
sudo nixos-rebuild switch --flake .#myhost
```

### Options
- `enable`: Enable the timer service.
- `schedule`: Systemd calendar expression (see `man systemd.time`).
- `user`/`group`: User/group to run as (created if not present).
- `persistent`: Run missed events after startup (default: true).
- `randomizedDelaySec`: Add random delay to timer (default: 0).

**easy-dca Configuration Options:**
- `publicKeyPath`/`privateKeyPath`: Paths to files containing Kraken API keys (**required**).
- `priceFactor`: Price factor for limit orders (default: 0.998).
- `fiatAmountPerBuy`: Fixed fiat amount per buy in EUR (optional).
- `monthlyFiatSpending`: Monthly fiat spending in EUR (optional, used if fiatAmountPerBuy is not set).
- `autoAdjustMinOrder`: Auto-adjust orders below minimum size (default: false).
- `dryRun`: Only validate orders (default: true).
- `notifyMethod`: Notification method (e.g., "ntfy").
- `notifyNtfyTopic`: ntfy topic (if using ntfy).
- `notifyNtfyURL`: ntfy server URL (if using ntfy).

**Legacy Options:**
- `environment`: Additional environment variables (legacy option).
- `extraArgs`: Additional command line arguments.

### Security
The service runs with strong systemd hardening by default (see flake.nix for details).

## License
Please see the LICENSE file for details.

## Example Config Files

- `.env.example`: Template for environment variables. Copy to `.env` and fill in your values.
- `docker-compose.yml`: Docker Compose configuration with default settings for production deployment.
- `examples/public.key.example`, `examples/private.key.example`: Example key files for use with Docker secrets or NixOS credentials. Replace with your real keys in production.
- `scripts/`: Utility scripts for running the application in different modes.

### Getting Started

```sh
# 1. Copy and configure environment variables
cp .env.example .env
# Edit .env with your real values

# 2. For Docker secrets (optional)
cp examples/public.key.example examples/public.key
cp examples/private.key.example examples/private.key
# Edit these files with your real API keys

# 3. Run the application
./scripts/test-once.sh  # Run once
./scripts/test-cron.sh  # Run with cron (set EASY_DCA_CRON in .env)
```

**Important:** Always edit the copied files with your real values before deploying.

## Transparency & AI Involvement

This project was developed with the assistance of large language model (LLM) coding agents. Automated code suggestions, refactoring, and documentation were generated and reviewed as part of the development process. Please review and audit the code for your own use case and security requirements. 

> **Note:** For ntfy notifications, both `NOTIFY_NTFY_TOPIC` and `NOTIFY_NTFY_URL` must be set. If `NOTIFY_NTFY_URL` is missing, notifications will be disabled and a warning will be logged.

### Scheduler Modes

The app supports different scheduling modes for different deployment scenarios:

1. **`manual` (default when no cron)**: Run once and exit - perfect for systemd timers, Docker one-shot containers, or manual execution
2. **`cron` (default when EASY_DCA_CRON is set)**: Use internal cron scheduling - good for standalone deployments or Docker containers that need to run continuously
3. **`systemd`**: Optimized for systemd timer integration - runs once and exits, letting systemd handle the scheduling

**When to use each mode:**

- **`manual`**: NixOS systemd timers, Kubernetes CronJobs, manual execution, CI/CD pipelines
- **`cron`**: Docker containers running continuously, standalone servers, when you want the app to handle its own scheduling
- **`systemd`**: NixOS systemd timers (explicit mode), when you want to be explicit about systemd integration

**Examples:**
```bash
# Manual mode (runs once)
EASY_DCA_SCHEDULER_MODE=manual

# Cron mode (internal scheduling)
EASY_DCA_SCHEDULER_MODE=cron
EASY_DCA_CRON="0 8 * * *"

# Systemd mode (for systemd timers)
EASY_DCA_SCHEDULER_MODE=systemd
```