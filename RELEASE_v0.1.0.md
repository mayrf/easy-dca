# 🎉 easy-dca v0.1.0 Release

## 🚀 What's New
- **Modular Architecture**: Separated scheduler, DCA runner, and notifications into distinct packages for better maintainability
- **Multiple Scheduler Modes**: Support for cron (internal), systemd timer, and manual execution modes
- **Enhanced Minimum Order Handling**: Runtime warnings and configurable auto-adjustment for orders below Kraken's minimum size
- **Improved Docker Support**: Fixed Dockerfile for Chainguard static images with better security and smaller size
- **Better Secret Management**: Support for file-based API key configuration via Docker secrets or NixOS credentials

## 🔧 Features
- **Scheduler Modes**: Choose between `cron` (internal scheduling), `systemd` (systemd timer integration), or `manual` (run once)
- **Minimum Order Size Warnings**: Get notified when orders are close to Kraken's minimum (0.00005 BTC)
- **Auto-adjustment Option**: Configurable behavior for orders below minimum size with `EASY_DCA_AUTO_ADJUST_MIN_ORDER`
- **Enhanced Notifications**: Improved ntfy integration with better error handling and validation
- **Docker Optimization**: Smaller, more secure images using Chainguard static builds
- **Flexible Buy Amount Configuration**: Support for both fixed amounts per buy and monthly spending targets
- **Price Factor Strategy**: Configurable limit order placement for better fees and prices

## 🛠️ Configuration
- `EASY_DCA_SCHEDULER_MODE`: Set to "cron", "systemd", or "manual" (default: "cron" if EASY_DCA_CRON is set, otherwise "manual")
- `EASY_DCA_AUTO_ADJUST_MIN_ORDER`: Enable/disable automatic order size adjustment (default: false)
- `EASY_DCA_PUBLIC_KEY_PATH` / `EASY_DCA_PRIVATE_KEY_PATH`: File-based secret management
- `EASY_DCA_FIAT_AMOUNT_PER_BUY`: Fixed fiat amount per buy (takes precedence over monthly spending)
- `EASY_DCA_MONTHLY_FIAT_SPENDING`: Monthly spending target (divided by number of buys per month)

## 📦 Installation
```bash
# Docker (GitHub Container Registry)
docker pull ghcr.io/mayrf/easy-dca:0.1.0

# Docker (build locally)
git clone https://github.com/mayrf/easy-dca.git
cd easy-dca
git checkout v0.1.0
docker build -t easy-dca .

# Local development
go build -o easy-dca ./cmd/easy-dca
```

## 🔄 Migration Notes
- **Backward Compatible**: Existing cron configurations will continue to work without changes
- **New Defaults**: `EASY_DCA_SCHEDULER_MODE` defaults to "cron" for backward compatibility
- **Safety First**: `EASY_DCA_AUTO_ADJUST_MIN_ORDER` defaults to `false` for safety
- **Optional Features**: New features are opt-in and won't affect existing setups

## 📋 Changelog
- feat: major architectural improvements and enhanced configurability
- fix: dockerfile not building
- chore: remove Codecov upload and badge from CI and docs
- fix: check notifier.Notify errors and remove unused getEnvAsInt (linter compliance)
- ci: add golangci-lint, Codecov coverage, Docker non-root user, and update docs with badge and CI info
- chore: use file paths for credentials in .env.example
- Reorganize project structure: move test scripts to scripts/ directory and simplify configuration approach
- chore: remove docker-compose.yaml
- Improve order response parsing and logging, require ntfy URL, fix config test, and update Dockerfile guidance
- Fix: dry-run logic bug

## 🔗 Links
- [Documentation](https://github.com/mayrf/easy-dca#readme)
- [Issues](https://github.com/mayrf/easy-dca/issues)
- [Docker Hub](https://ghcr.io/mayrf/easy-dca)
- [Source Code](https://github.com/mayrf/easy-dca)

## ⚠️ Important Notes
- **First Stable Release**: This is the first stable release (v0.1.0) - test thoroughly before using with real funds
- **Dry Run Recommended**: Always start with `EASY_DCA_DRY_RUN=true` to validate your configuration
- **API Limits**: Be aware of Kraken API rate limits and minimum order sizes
- **Security**: Use file-based secrets when possible, especially in production environments
- **Testing**: The modular architecture makes it easier to test individual components

## 🎯 Release Highlights
- **Architecture**: Clean separation of concerns with modular design
- **Flexibility**: Multiple scheduling options to fit different deployment scenarios
- **Safety**: Enhanced error handling and minimum order size management
- **Security**: Improved Docker security with Chainguard images and non-root user
- **Usability**: Better configuration options and documentation

## 🔮 Future Roadmap
- Additional notification providers (email, Slack, Discord)
- More exchange integrations
- Web UI for configuration and monitoring
- Advanced order types (stop-loss, take-profit)
- Portfolio tracking and analytics

---

**Thank you for using easy-dca! 🚀**

This release represents a significant milestone in the project's development, providing a stable foundation for automated DCA trading with enhanced safety and flexibility features. 