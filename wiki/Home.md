# GOVARD: Go-based Versatile Runtime & Development

[![Go Version](https://img.shields.io/github/go-mod/go-version/ddtcorex/govard)](https://go.dev/)
[![License](https://img.shields.io/github/license/ddtcorex/govard)](LICENSE)
[![Releases](https://img.shields.io/github/v/release/ddtcorex/govard)](https://github.com/ddtcorex/govard/releases)
[![CI Pipeline](https://github.com/ddtcorex/govard/actions/workflows/ci-pipeline.yml/badge.svg)](https://github.com/ddtcorex/govard/actions/workflows/ci-pipeline.yml)

**Govard** is a professional-grade local development orchestrator engineered in Go. It replaces legacy bash-based tools with a high-performance native binary that manages complex containerized environments with a focus on stability, speed, and a premium developer experience.

---

## 🆚 Why Govard Stands Out

| Area | Govard Advantage |
| :--- | :--- |
| **Core architecture** | Native Go binary with direct Docker SDK orchestration for predictable lifecycle behavior |
| **Framework intelligence** | Automatic framework discovery + framework-specific blueprints + custom stack wizard |
| **Magento depth** | First-class Magento 2 workflow with auto `env.php` wiring, optional Varnish/Redis/queue/search |
| **Local HTTPS/DNS** | Built-in Caddy + `dnsmasq` + Root CA auto-trust for `*.test` domains |
| **Remote safety** | `remote`/`sync` protections with `prod` write blocking, scoped capabilities, audit logs |
| **Team reproducibility** | `govard lock` + `lock.strict` to detect environment drift across machines |
| **Recovery workflow** | `govard snapshot` for quick local DB/media checkpoints |
| **CLI + Desktop parity** | Same core engine in both CLI and Wails Desktop app |
| **Update integrity** | `govard self-update` validates release checksums before replacing binaries |

---

## 🚀 Key Features

- **🐳 Docker SDK Orchestration** — Direct Docker SDK integration, no shell-script glue
- **🔍 Framework Auto-Detection** — Detects Magento 1/2, Laravel, Next.js, Emdash, Drupal, Symfony, Shopware, CakePHP, WordPress
- **🔒 Local HTTPS** — Caddy proxy + dnsmasq + Root CA auto-trust for `*.test` domains
- **🌐 Remote Management** — Named remotes with scoped capabilities, SSH, sync, and audit logs
- **💾 Database Tools** — Dump, import, query, live monitoring (`db top`), privacy filters (`--no-pii`, `--no-noise`)
- **📸 Snapshots** — DB/media snapshots with gzip compression and bidirectional remote transfer
- **🐛 Zero-Config Debugging** — Xdebug 2 & 3 with one-click toggling and dedicated `php-debug` routing
- **🖥️ Desktop App** — Wails-based GUI with live logs, quick actions, and shell launcher
- **⬆️ Framework Upgrades** — Native upgrade pipeline for Magento 2, Laravel, Symfony, WordPress
- **🔄 Self-Update** — Checksum-verified binary replacement

---

## 📚 Documentation

| Page | Description |
| :--- | :--- |
| [Installation](Installation) | Install Govard on Linux or macOS |
| [Getting Started](Getting-Started) | First project walkthrough and daily workflow |
| [CLI Commands](CLI-Commands) | Full CLI reference with all commands and flags |
| [Configuration](Configuration) | `.govard.yml` structure, profiles, blueprints |
| [Frameworks](Frameworks) | Support matrix, runtime defaults, version-aware profiles |
| [Remotes and Sync](Remotes-and-Sync) | Remote setup, sync flows, snapshots, DB workflows |
| [SSL and Domains](SSL-and-Domains) | HTTPS, DNS, CA trust, domain management |
| [Desktop App](Desktop-App) | GUI surface, dev mode, frontend layout |
| [Architecture](Architecture) | System design, modules, extension points |
| [Contributing](Contributing) | Build, test, and contribution workflow |
| [FAQ & Troubleshooting](FAQ) | Common issues and solutions |
| [Changelog](Changelog) | Release history |

---

## ⚡ Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash

# Initialize a project
cd /path/to/your/project
govard init

# Start the environment
govard env up

# Enter the workspace
govard shell
```

→ **[Full installation guide](Installation)** | **[Getting Started guide](Getting-Started)**
