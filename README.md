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
- NixOS module for secure systemd timer integration

## Quick Start

### Docker Compose (Recommended for most users)

#### 1. Clone the repository
```sh
git clone https://github.com/mayrf/easy-dca.git
cd easy-dca
```

#### 2. Configure your environment
Copy the example environment file and customize it:
```sh
cp .env.example .env
```

#### 3. Configure your API keys
For Docker secrets (recommended):
```sh
cp examples/public.key.example examples/public.key
cp examples/private.key.example examples/private.key
# Edit these files with your real API keys
```

#### 4. Build and run
```sh
# Run once (manual mode)
./scripts/test-once.sh

# Run with cron flag
./scripts/test-cron-flag.sh

# Or use docker compose
docker-compose up
```

**Note:** After the first build, if you make changes to the Go code, you'll need to rebuild:
```sh
docker-compose up --build
```

The app will run on the schedule you set in `EASY_DCA_CRON` and send notifications via ntfy.

**Tip:** When the application starts, it will display a comprehensive summary of your configuration, explaining what each setting does in plain English.

### NixOS (For NixOS users)

If you use NixOS with flakes, you can enable `easy-dca` as a secure systemd timer service:

#### 1. Add the Flake as an Input
```nix
{
  inputs.easy-dca.url = "github:mayrf/easy-dca";
  # ... other inputs ...
}
```

#### 2. Import and Configure
```nix
{
  outputs = { self, nixpkgs, easy-dca, ... }@inputs: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        # ... other modules ...
        easy-dca.nixosModules.easy-dca
        ({ config, pkgs, ... }: {
          services.easy-dca = {
            enable = true;
            schedule = "*-*-* 02:30:00"; # 2:30am every day
            user = "easy-dca";
            group = "easy-dca";
            
            # Required API keys (using systemd credentials)
            publicKeyPath = "/run/secrets/kraken-public-key";
            privateKeyPath = "/run/secrets/kraken-private-key";
            
            # Trading configuration
            priceFactor = 0.998;
            dryRun = true; # Only validate orders (default)
            autoAdjustMinOrder = false;
            
            # Buy amount configuration
            fiatAmountPerBuy = 10.0; # Fixed amount per buy (required for systemd timers)
            
            # Notification configuration (optional)
            notifyMethod = "ntfy";
            notifyNtfyTopic = "yourtopic";
            notifyNtfyURL = "https://ntfy.sh";
          };
        })
      ];
    };
  };
}
```

#### 3. Set up secrets (with sops-nix or similar)
```nix
sops.secrets = {
  "kraken-public-key" = {
    path = "/run/secrets/kraken-public-key";
  };
  "kraken-private-key" = {
    path = "/run/secrets/kraken-private-key";
  };
};
```

#### 4. Rebuild and switch
```sh
sudo nixos-rebuild switch --flake .#myhost
```

## Configuration

### Environment Variables

#### Required
- `EASY_DCA_PUBLIC_KEY`: Kraken API public key
- `EASY_DCA_PRIVATE_KEY`: Kraken API private key

**Getting Kraken API Keys:**
- [How to Create a Kraken API Key](https://support.kraken.com/articles/360000919966-how-to-create-an-api-key)
- [Kraken API Documentation](https://docs.kraken.com/api/docs/rest-api/add-order)

**Required API Permissions:** `Orders and trades - Create & modify orders`

#### Trading Configuration
- `EASY_DCA_PAIR`: Trading pair (default: "BTC/EUR"). Supported pairs: BTC/EUR, BTC/GBP, BTC/CHF, BTC/AUD, BTC/CAD, BTC/USD
- `EASY_DCA_PRICE_FACTOR`: Price factor for limit orders (default: 0.998)
- `EASY_DCA_MONTHLY_FIAT_SPENDING`: Monthly fiat spending (optional, used if EASY_DCA_FIAT_AMOUNT_PER_BUY is not set)
- `EASY_DCA_FIAT_AMOUNT_PER_BUY`: Fixed fiat amount to spend each run (optional, takes precedence over EASY_DCA_MONTHLY_FIAT_SPENDING)
- `EASY_DCA_AUTO_ADJUST_MIN_ORDER`: If true, automatically adjust orders below minimum size (0.00005 BTC); if false, let them fail (default: false)
- `EASY_DCA_DRY_RUN`: If true (default), only validate orders (dry run); if false, actually place orders
- `EASY_DCA_DISPLAY_SATS`: If true, display all BTC amounts in satoshi (default: false)

#### Scheduling
- `EASY_DCA_CRON`: Cron expression for scheduling (optional; if not set, runs once)
- `EASY_DCA_SCHEDULER_MODE`: Scheduler mode: "cron", "systemd", or "manual" (default: "cron" if EASY_DCA_CRON is set, otherwise "manual")

#### Notifications
- `NOTIFY_METHOD`: Notification method (e.g., `ntfy`)
- `NOTIFY_NTFY_TOPIC`: ntfy topic (if using ntfy)
- `NOTIFY_NTFY_URL`: ntfy server URL (**required for ntfy notifications**; no default)

#### Logging
- `EASY_DCA_LOG_FORMAT`: Log format control (default: no timestamp)
  - `"timestamp"` or `"time"`: Standard format (2006/01/02 15:04:05)
  - `"microseconds"` or `"micro"`: Full datetime with microseconds (2006/01/02 15:04:05.000000)
  - Any other value or unset: No timestamp prefix

### Secret Management

For better security, you can provide your Kraken API keys via file paths instead of directly in environment variables:

- `EASY_DCA_PUBLIC_KEY_PATH`: Path to a file containing your Kraken API public key
- `EASY_DCA_PRIVATE_KEY_PATH`: Path to a file containing your Kraken API private key

This is preferred because secrets are not exposed in environment variables and integrates well with Docker secrets, NixOS systemd credentials, and other secret managers.

### NixOS Module Options

When using the NixOS module, you can configure the service using these options:

#### Service Configuration
- `enable`: Enable the timer service
- `schedule`: Systemd calendar expression for when to run the service (e.g., `"*-*-* 02:30:00"` for 2:30am daily)
- `user`/`group`: User/group to run as (created if not present)
- `persistent`: Run missed events after startup (default: true)
- `randomizedDelaySec`: Add random delay to timer (default: 0)

#### Trading Configuration
- `publicKeyPath`/`privateKeyPath`: Paths to files containing Kraken API keys (**required**)
- `priceFactor`: Price factor for limit orders (default: 0.998)
- `fiatAmountPerBuy`: Fixed fiat amount per buy in fiat currency (**required for systemd timers**)
- `monthlyFiatSpending`: Monthly fiat spending in fiat currency (**not available with systemd timers** - see limitations below)
- `autoAdjustMinOrder`: Auto-adjust orders below minimum size (default: false)
- `dryRun`: Only validate orders (default: true)
- `displaySats`: Display BTC amounts in sats (default: false)

#### Notification Configuration
- `notifyMethod`: Notification method (e.g., "ntfy")
- `notifyNtfyTopic`: ntfy topic (if using ntfy)
- `notifyNtfyURL`: ntfy server URL (if using ntfy)

#### How Scheduling Works
- Set your schedule using the `schedule` option with systemd calendar format (e.g., `"*-*-* 02:30:00"` for 2:30am daily)
- The module uses this directly as the systemd OnCalendar string for the timer
- **Important**: When using systemd timers, you **must** use `fiatAmountPerBuy` for buy amounts
- The `monthlyFiatSpending` option is **not available** with systemd timers because the application cannot reliably calculate the number of executions per month from the systemd timer schedule

#### Systemd Timer Limitations
- **Monthly buy calculations don't work**: The application cannot determine how many times per month the systemd timer will execute
- **Use fixed amounts only**: Set `fiatAmountPerBuy` to specify exactly how much to spend each time the timer runs
- **No cron expression parsing**: The systemd timer handles scheduling, not the application's cron parser

## Trading Strategy

### Supported Trading Pairs

The application supports the following BTC trading pairs:
- **BTC/EUR** - Bitcoin/Euro (default)
- **BTC/GBP** - Bitcoin/British Pound
- **BTC/CHF** - Bitcoin/Swiss Franc
- **BTC/AUD** - Bitcoin/Australian Dollar
- **BTC/CAD** - Bitcoin/Canadian Dollar
- **BTC/USD** - Bitcoin/US Dollar

### Buy Amount Configuration

The app supports two ways to configure how much to buy:

1. **Fixed amount per buy**: Set `EASY_DCA_FIAT_AMOUNT_PER_BUY` to spend the same fiat amount each time
2. **Monthly amount**: Set `EASY_DCA_MONTHLY_FIAT_SPENDING` and the app will divide it by the number of buys per month (calculated from your cron schedule)

**Examples:**
- **Spend 10 EUR every day**: Set `EASY_DCA_PAIR="BTC/EUR"`, `EASY_DCA_FIAT_AMOUNT_PER_BUY=10` and `EASY_DCA_CRON="0 8 * * *"`
- **Spend 300 EUR per month, buying every 3 days**: Set `EASY_DCA_PAIR="BTC/EUR"`, `EASY_DCA_MONTHLY_FIAT_SPENDING=300` and `EASY_DCA_CRON="0 8 */3 * *"` (app will spend ~30 EUR each time)
- **Spend 150 EUR per month, buying weekly**: Set `EASY_DCA_PAIR="BTC/EUR"`, `EASY_DCA_MONTHLY_FIAT_SPENDING=150` and `EASY_DCA_CRON="0 8 * * 1"` (app will spend ~37.5 EUR each time)
- **Spend 20 USD daily**: Set `EASY_DCA_PAIR="BTC/USD"`, `EASY_DCA_FIAT_AMOUNT_PER_BUY=20` and `EASY_DCA_CRON="0 8 * * *"`

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

The `EASY_DCA_PRICE_FACTOR` determines at what percentage of the current best sell offer (ask price) your buy orders are placed:

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

## Scheduler Modes

The app supports different scheduling modes for different deployment scenarios:

1. **`manual` (default when no cron)**: Run once and exit - perfect for systemd timers, Docker one-shot containers, or manual execution
2. **`cron` (default when EASY_DCA_CRON is set)**: Use internal cron scheduling - good for standalone deployments or Docker containers that need to run continuously
3. **`systemd`**: Optimized for systemd timer integration - runs once and exits, letting systemd handle the scheduling. **Note**: Monthly buy amount calculations (`EASY_DCA_MONTHLY_FIAT_SPENDING`) are not supported in this mode; use `EASY_DCA_FIAT_AMOUNT_PER_BUY` instead.

**When to use each mode:**
- **`manual`**: NixOS systemd timers, Kubernetes CronJobs, manual execution, CI/CD pipelines
- **`cron`**: Docker containers running continuously, standalone servers, when you want the app to handle its own scheduling
- **`systemd`**: NixOS systemd timers (explicit mode), when you want to be explicit about systemd integration. **Requires fixed buy amounts only.**

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

## Development

### CI/CD
- GitHub Actions workflow runs linting, tests, and builds the Docker image on every push and pull request to `master`
- Linting is performed using `golangci-lint` to ensure code quality

### Extending Notifications
To add more notification backends (Slack, Email, etc.), implement the `Notifier` interface in `cmd/easy-dca/easy-dca.go` and add a case to `getNotifier()`

### Example Config Files
- `.env.example`: Template for environment variables. Copy to `.env` and fill in your values
- `docker-compose.yml`: Docker Compose configuration with default settings for production deployment
- `examples/public.key.example`, `examples/private.key.example`: Example key files for use with Docker secrets or NixOS credentials. Replace with your real keys in production
- `scripts/`: Utility scripts for running the application in different modes

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

## Security

The NixOS service runs with strong systemd hardening by default (see flake.nix for details).

## License

Please see the LICENSE file for details.

## Transparency & AI Involvement

This project was developed with the assistance of large language model (LLM) coding agents. Automated code suggestions, refactoring, and documentation were generated and reviewed as part of the development process. Please review and audit the code for your own use case and security requirements.

> **Note:** For ntfy notifications, both `NOTIFY_NTFY_TOPIC` and `NOTIFY_NTFY_URL` must be set. If `NOTIFY_NTFY_URL` is missing, notifications will be disabled and a warning will be logged.

**Pro Tip:** If you have the [Kraken Pro app](https://www.kraken.com/features/cryptocurrency-apps) installed on your phone, you'll receive a notification once your DCA order has been filled!
