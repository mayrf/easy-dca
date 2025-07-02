# easy-dca

A Go application for automated DCA (Dollar Cost Averaging) trading on Kraken, with cron scheduling and pluggable notifications (ntfy, etc).

## Features
- Run once or on a schedule (cron expression via CLI flag or env var)
- Pluggable notification system (ntfy supported, extensible for others)
- Simple configuration via environment variables
- Ready for Docker and docker-compose deployment (uses Chainguard images for security and minimal size)
- CI/CD with GitHub Actions

## Quick Start (Docker Compose)

### 1. Clone the repository
```sh
git clone https://github.com/mayrf/easy-dca.git
cd easy-dca
```

### 2. Create a `.env` file
Copy the example and fill in your values:
```sh
cp .env.example .env
# Edit .env with your actual values
```

### 3. Create a `docker-compose.yml` file
Copy the example:
```sh
cp docker-compose.example.yml docker-compose.yml
```

### 4. Build and run
```sh
docker-compose up --build
```

The app will run on the schedule you set in `EASY_DCA_CRON` and send notifications via ntfy.

## Quick Start (Local Development)

### 1. Clone the repository
```sh
git clone https://github.com/mayrf/easy-dca.git
cd easy-dca
```

### 2. Create a `.env` file
```sh
cp .env.example .env
# Edit .env with your actual values
```

### 3. Run the application
```sh
# Run once
./scripts/test-once.sh

# Run with cron (set EASY_DCA_CRON in .env)
./scripts/test-cron.sh

# Run with cron flag
./scripts/test-cron-flag.sh
```

## Environment Variables
- `EASY_DCA_PUBLIC_KEY`: Kraken API public key (**required**)
- `EASY_DCA_PRIVATE_KEY`: Kraken API private key (**required**)
- `EASY_DCA_PRICEFACTOR`: Price factor for limit orders (default: 0.998)
- `EASY_DCA_MONTHLY_VOLUME`: Monthly trading volume (default: 140.0)
- `EASY_DCA_CRON`: Cron expression for scheduling (optional; if not set, runs once)
- `NOTIFY_METHOD`: Notification method (e.g., `ntfy`)
- `NOTIFY_NTFY_TOPIC`: ntfy topic (if using ntfy)
- `NOTIFY_NTFY_URL`: ntfy server URL (**required for ntfy notifications**; no default)
- `EASY_DCA_DRY_RUN`: If true (default), only validate orders (dry run); if false, actually place orders.

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
- GitHub Actions workflow runs tests and builds the Docker image on every push and pull request to `master`.

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
            environment = {
              EASY_DCA_PUBLIC_KEY = "your_public_key";
              EASY_DCA_PRIVATE_KEY = "your_private_key";
              EASY_DCA_PRICEFACTOR = "0.998";
              EASY_DCA_MONTHLY_VOLUME = "140.0";
              NOTIFY_METHOD = "ntfy";
              NOTIFY_NTFY_TOPIC = "yourtopic";
              EASY_DCA_DRY_RUN = "true"; # Only validate orders (default)
              # EASY_DCA_DRY_RUN = "false"; # Actually place orders
              # NOTIFY_NTFY_URL = "https://ntfy.sh"; # Optional
            };
            # credentials = {
            #   API_KEY = "/run/secrets/api-key";
            # };
            # extraArgs = [ ];
          };
        })
      ];
    };
  };
}
```

### 3. Rebuild and Switch
```sh
sudo nixos-rebuild switch --flake .#myhost
```

### Options
- `enable`: Enable the timer service.
- `schedule`: Systemd calendar expression (see `man systemd.time`).
- `user`/`group`: User/group to run as (created if not present).
- `environment`: Environment variables for the app.
- `credentials`: Map env vars to secret file paths (securely loaded).
- `extraArgs`: Additional CLI arguments for the app.
- `persistent`: Run missed events after startup (default: true).
- `randomizedDelaySec`: Add random delay to timer (default: 0).

### Security
The service runs with strong systemd hardening by default (see flake.nix for details).

## License
Please see the LICENSE file for details.

## Example Config Files

- `.env.example`: Template for environment variables. Copy to `.env` and fill in your values.
- `docker-compose.example.yml`: Reference Compose file showing best practices for secrets and env config.
- `examples/public.key.example`, `examples/private.key.example`: Example key files for use with Docker secrets or NixOS credentials. Replace with your real keys in production.
- `scripts/`: Utility scripts for running the application in different modes.

### Getting Started

```sh
# 1. Copy and configure environment variables
cp .env.example .env
# Edit .env with your real values

# 2. For Docker Compose (optional)
cp docker-compose.example.yml docker-compose.yml

# 3. For Docker secrets (optional)
cp examples/public.key.example examples/public.key
cp examples/private.key.example examples/private.key
# Edit these files with your real API keys

# 4. Run the application
./scripts/test-once.sh  # Run once
./scripts/test-cron.sh  # Run with cron (set EASY_DCA_CRON in .env)
```

**Important:** Always edit the copied files with your real values before deploying.

## Transparency & AI Involvement

This project was developed with the assistance of large language model (LLM) coding agents. Automated code suggestions, refactoring, and documentation were generated and reviewed as part of the development process. Please review and audit the code for your own use case and security requirements. 

> **Note:** For ntfy notifications, both `NOTIFY_NTFY_TOPIC` and `NOTIFY_NTFY_URL` must be set. If `NOTIFY_NTFY_URL` is missing, notifications will be disabled and a warning will be logged.
