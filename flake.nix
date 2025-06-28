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

        # Go application build
        goApp = pkgs.buildGoModule {
          pname = "easy-dca";
          version = "0.1.0";

          src = ./.;

          # You'll need to update this hash after first build attempt
          # Run `nix build` and it will tell you the correct hash
          vendorHash = "sha256-NHTKwUSIbNCUco88JbHOo3gt6S37ggee+LWNbHaRGEs=";

          # Optional: specify Go version if needed
          # buildInputs = [ pkgs.go_1_21 ];

          meta = with pkgs.lib; {
            description = "My Go application";
            homepage = "https://github.com/username/my-go-app";
            license = licenses.mit;
            maintainers = [ ];
          };
        };
      in {
        # Default package (the built Go application)
        packages.default = goApp;

        # You can also expose it with a specific name
        packages.my-go-app = goApp;

        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls # Go language server
            gotools # goimports, godoc, etc.
            go-migrate # Database migrations (optional)
            delve # Go debugger (optional)
            air # Live reload for Go apps (optional)
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
      });
}
