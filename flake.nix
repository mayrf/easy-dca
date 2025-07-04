{
  description = "Easy dca go flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        easy-dca = pkgs.buildGoModule {
          pname = "easy-dca";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-QFHoPpewvzRqsk1XhsheKScl5pOQR3RU0ZYGkTedY8s=";
          subPackages = [ "cmd/easy-dca" ];

          # Optional: specify Go version if needed
          buildInputs = [ pkgs.go_1_24 ];

          meta = with pkgs.lib; {
            description = "easy-dca application";
            homepage = "https://github.com/mayrf/easy-dca";
            license = licenses.gpl3;
            maintainers = [ ];
          };
        };
      in {
        packages.default = easy-dca;
        packages.easy-dca = easy-dca;

        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls # Go language server
            gotools # goimports, godoc, etc.
            # delve # Go debugger (optional)
          ];

          # Environment variables for development
          shellHook = ''
            echo "Go development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go run .     - Run the application"
            echo "  go test ./... - Run tests"
            echo "  go build .   - Build the application"
            echo "  nix build    - Build with Nix"
          '';
        };

        # Optional: provide a formatter for `nix fmt`
        formatter = pkgs.nixpkgs-fmt;
      }) // {
        # NixOS module (system-independent)
        nixosModules.easy-dca = { config, lib, pkgs, ... }:
          with lib;
          let
            cfg = config.services.easy-dca;

            # Get the Go application from the flake
            easy-dca-app = self.packages.${pkgs.system}.easy-dca;

          in {
            options.services.easy-dca = {
              enable = mkEnableOption "easy-dca Timer Service";

              schedule = mkOption {
                type = types.str;
                default = "daily";
                description =
                  "Systemd calendar expression for when to run the service";
                example = "*-*-* 02:30:00";
              };

              user = mkOption {
                type = types.str;
                default = "nobody";
                description = "User to run the service as";
              };

              group = mkOption {
                type = types.str;
                default = "nogroup";
                description = "Group to run the service as";
              };

              persistent = mkOption {
                type = types.bool;
                default = true;
                description =
                  "Whether to run missed timer events after system startup";
              };

              randomizedDelaySec = mkOption {
                type = types.str;
                default = "0";
                description =
                  "Random delay before execution (e.g., '30m', '1h')";
              };

              # easy-dca specific configuration options
              publicKeyPath = mkOption {
                type = types.path;
                description = "Path to file containing Kraken API public key";
                example = "/run/secrets/kraken-public-key";
              };

              privateKeyPath = mkOption {
                type = types.path;
                description = "Path to file containing Kraken API private key";
                example = "/run/secrets/kraken-private-key";
              };

              priceFactor = mkOption {
                type = types.float;
                default = 0.998;
                description = "Price factor for limit orders (0.95-0.9999)";
                example = 0.998;
              };

              fiatAmountPerBuy = mkOption {
                type = types.nullOr types.float;
                default = null;
                description = "Fixed fiat amount in EUR to spend each run";
                example = 10.0;
              };

              monthlyFiatSpending = mkOption {
                type = types.nullOr types.float;
                default = null;
                description = "Monthly fiat spending in EUR (used if fiatAmountPerBuy is not set)";
                example = 300.0;
              };

              autoAdjustMinOrder = mkOption {
                type = types.bool;
                default = false;
                description = "Automatically adjust orders below minimum size (0.00005 BTC)";
              };

              dryRun = mkOption {
                type = types.bool;
                default = true;
                description = "Only validate orders (dry run); if false, actually place orders";
              };

              notifyMethod = mkOption {
                type = types.nullOr types.str;
                default = null;
                description = "Notification method (e.g., 'ntfy')";
                example = "ntfy";
              };

              notifyNtfyTopic = mkOption {
                type = types.nullOr types.str;
                default = null;
                description = "ntfy topic (if using ntfy)";
                example = "your_ntfy_topic";
              };

              notifyNtfyURL = mkOption {
                type = types.nullOr types.str;
                default = null;
                description = "ntfy server URL (if using ntfy)";
                example = "https://ntfy.sh";
              };

              # Legacy options for backward compatibility
              environment = mkOption {
                type = types.attrsOf types.str;
                default = { };
                description = "Additional environment variables for the service (legacy option)";
                example = {
                  LOG_LEVEL = "info";
                  DATA_DIR = "/var/lib/myapp";
                };
              };

              extraArgs = mkOption {
                type = types.listOf types.str;
                default = [ ];
                description = "Additional command line arguments";
                example = [ "--config" "/etc/myapp/config.yaml" ];
              };
            };

            config = mkIf cfg.enable {
              # Create the systemd timer
              systemd.timers."easy-dca-timer-service" = {
                wantedBy = [ "timers.target" ];
                timerConfig = {
                  OnCalendar = cfg.schedule;
                  Persistent = cfg.persistent;
                  RandomizedDelaySec = cfg.randomizedDelaySec;
                };
              };

              # Create the systemd service
              systemd.services."easy-dca-timer-service" = {
                description = "easy-dca Timer Service";
                serviceConfig = {
                  Type = "oneshot";
                  User = cfg.user;
                  Group = cfg.group;
                  ExecStart = "${easy-dca-app}/bin/easy-dca ${
                      concatStringsSep " " cfg.extraArgs
                    }";

                  # Security hardening
                  NoNewPrivileges = true;
                  ProtectSystem = "strict";
                  ProtectHome = true;
                  PrivateTmp = true;
                  ProtectKernelTunables = true;
                  ProtectKernelModules = true;
                  ProtectControlGroups = true;
                  ProtectProc = "invisible";
                  RestrictAddressFamilies = [ "AF_INET" "AF_INET6" ];
                  CapabilityBoundingSet = "";
                  PrivateDevices = true;
                  ProtectClock = true;
                  ProtectHostname = true;
                  ProtectKernelLogs = true;
                  RestrictRealtime = true;
                  SystemCallFilter = [ "@system-service" ];
                  LoadCredential = [
                    "kraken-public-key:${cfg.publicKeyPath}"
                    "kraken-private-key:${cfg.privateKeyPath}"
                  ];
                };

                # Set environment variables from module options
                environment = cfg.environment // (let
                  # Build conditional environment variables
                  conditionalEnv = {}
                    // (if cfg.fiatAmountPerBuy != null then { EASY_DCA_FIAT_AMOUNT_PER_BUY = toString cfg.fiatAmountPerBuy; } else {})
                    // (if cfg.monthlyFiatSpending != null then { EASY_DCA_MONTHLY_FIAT_SPENDING = toString cfg.monthlyFiatSpending; } else {})
                    // (if cfg.notifyMethod != null then { NOTIFY_METHOD = cfg.notifyMethod; } else {})
                    // (if cfg.notifyNtfyTopic != null then { NOTIFY_NTFY_TOPIC = cfg.notifyNtfyTopic; } else {})
                    // (if cfg.notifyNtfyURL != null then { NOTIFY_NTFY_URL = cfg.notifyNtfyURL; } else {});
                in {
                  # Required API keys (using systemd credentials)
                  EASY_DCA_PUBLIC_KEY_PATH = "%d/kraken-public-key";
                  EASY_DCA_PRIVATE_KEY_PATH = "%d/kraken-private-key";
                  
                  # Trading configuration
                  EASY_DCA_PRICEFACTOR = toString cfg.priceFactor;
                  EASY_DCA_DRY_RUN = toString cfg.dryRun;
                  EASY_DCA_AUTO_ADJUST_MIN_ORDER = toString cfg.autoAdjustMinOrder;
                  
                  # Scheduler mode (always systemd for NixOS)
                  EASY_DCA_SCHEDULER_MODE = "systemd";
                } // conditionalEnv);
              };

              # Ensure the user exists if it's not a system user
              users.users = mkIf (cfg.user != "nobody" && cfg.user != "root") {
                ${cfg.user} = {
                  isSystemUser = true;
                  group = cfg.group;
                  description = "easy-dca Timer Service user";
                };
              };

              users.groups =
                mkIf (cfg.group != "nogroup" && cfg.group != "root") {
                  ${cfg.group} = { };
                };
            };
          };

        # Default NixOS module
        nixosModules.default = self.nixosModules.easy-dca;
      };
}
