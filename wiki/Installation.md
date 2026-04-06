# Installation

This page covers all methods to install Govard on Linux and macOS.

> [!IMPORTANT]
> Do not mix install channels on the same machine (e.g., `.deb` + `make install` + `self-update` across different paths). Use **one channel only** to avoid conflicting binaries in `/usr/bin` and `/usr/local/bin`.

---

## 🚀 One-Line Install (Linux/macOS)

Install the latest release binary with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash
```

Using `wget`:

```bash
wget -qO- https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash
```

### Common Install Options

```bash
# Install to ~/.local/bin (no sudo required)
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash -s -- --local

# Build from source (auto-installs Go 1.25 if needed)
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash -s -- --source
```

By default, this installs both `govard` and `govard-desktop` to `/usr/local/bin` and:
- Auto-detects/installs missing system dependencies (`certutil`, `WebKitGTK`)
- Starts global services
- Configures SSL trust
- On Linux, falls back to extracting `govard-desktop` from the `.deb` package if a standalone archive is not in the release

---

## 📦 Release Installers (CLI + Desktop)

Every tagged release publishes installer packages that include both `govard` (CLI) and `govard-desktop`.

From the [releases page](https://github.com/ddtcorex/govard/releases):

### Linux (`.deb`)

```bash
sudo dpkg -i govard_<version>_linux_amd64.deb
```

### macOS (`.pkg`)

```bash
sudo installer -pkg govard_<version>_Darwin_arm64.pkg -target /
```

---

## 🔧 Build from Source

### Prerequisites

Ensure you have the following installed:

| Tool | Required Version |
| :--- | :--- |
| Go | `1.25+` |
| Node.js | `20+` |
| Yarn | v1.x |
| golangci-lint | v2.11+ |
| Docker + Docker Compose | latest |
| Wails | `v2.11+` (desktop development only) |

### Source Install

```bash
git clone https://github.com/ddtcorex/govard.git
cd govard
./install.sh --source
```

### Local Developer Setup

1. **Install Go 1.25+** from [go.dev](https://go.dev/dl/)

2. **Enable Yarn** via Corepack:
   ```bash
   corepack enable
   ```

3. **Install golangci-lint**:
   ```bash
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

4. **Install Wails** (for desktop development):
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   wails version
   ```

No `sudo` required — install everything locally and update your `PATH`.

---

## 🐳 Docker Images

Govard uses a single PHP Dockerfile with build args instead of versioned folders.

```bash
# Standard PHP image
docker build -f docker/php/Dockerfile \
  -t ddtcorex/govard-php:8.4 \
  --build-arg PHP_VERSION=8.4 \
  docker/php

# Magento 2 optimized PHP image
docker build -f docker/php/magento2/Dockerfile \
  -t ddtcorex/govard-php-magento2:8.4 \
  --build-arg PHP_VERSION=8.4 \
  docker/php
```

---

## 🔄 Updating Govard

```bash
govard self-update
```

`self-update` downloads the platform-specific release artifact, **verifies the SHA-256 checksum**, and atomically replaces installed binaries (`govard` + `govard-desktop`).

---

## ✅ Verify Installation

```bash
govard version
govard doctor
```

`govard doctor` runs system diagnostics including Docker, DNS, ports, and SSL trust checks.

---

**[← Home](Home)** | **[Getting Started →](Getting-Started)**
