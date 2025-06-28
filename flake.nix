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

          vendorHash = "sha256-NHTKwUSIbNCUco88JbHOo3gt6S37ggee+LWNbHaRGEs=";

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

              environment = mkOption {
                type = types.attrsOf types.str;
                default = { };
                description = "Environment variables for the service";
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
              credentials = mkOption {
                type = types.attrsOf types.path;
                default = { };
                description = ''
                  Credentials to load as environment variables.
                  Maps environment variable names to paths of secret files.
                '';
                example = {
                  API_KEY = "/run/secrets/api-key";
                  DATABASE_PASSWORD = "/run/secrets/db-password";
                };
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
                  LoadCredential = mapAttrsToList
                    (envVar: secretPath: "${envVar}:${secretPath}")
                    cfg.credentials;

                  # Set environment variables that read from the loaded credentials
                  # systemd places loaded credentials in $CREDENTIALS_DIRECTORY/<name>
                  Environment =
                    mapAttrsToList (envVar: _: "${envVar}=%d/${envVar}")
                    cfg.credentials;
                };

                # Set environment variables
                environment = cfg.environment;
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
