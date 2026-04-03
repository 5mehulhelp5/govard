# Getting Started

This guide covers the shortest path from install to a working Govard project.

## Install Govard

### Release installer

```bash
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash
```

Useful variations:

```bash
# Install into ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash -s -- --local

# Build from source during install
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash -s -- --source
```

Tagged releases also publish installers that include both `govard` and `govard-desktop`:

- Linux: `govard_<version>_linux_<arch>.deb`
- macOS: `govard_<version>_Darwin_<arch>.pkg`

Do not mix install channels on one machine. Pick one path and keep it consistent.

### Local source setup

Minimum local toolchain:

- Go `1.25+`
- Node.js `20+`
- Docker Engine + Docker Compose plugin
- Wails `v2.11+` if you work on the desktop app

```bash
git clone https://github.com/ddtcorex/govard.git
cd govard
./install.sh --source
```

## First Project

### 1. Initialize

```bash
cd /path/to/project
govard init
```

Govard inspects `composer.json` or `package.json`, detects the framework, and writes `.govard.yml`.

Detected frameworks:

- Magento 1 / OpenMage
- Magento 2
- Laravel
- Next.js
- Drupal
- Symfony
- Shopware
- CakePHP
- WordPress
- Custom stack via interactive prompts

### 2. Start the environment

```bash
govard env up
```

Common variants:

```bash
govard up --quickstart
govard env up --pull
govard env up --fallback-local-build
```

The startup pipeline is:

1. Detect framework context
2. Validate config, Docker, ports, and runtime prerequisites
3. Render the compose file into `~/.govard/compose/`
4. Start containers
5. Verify proxy and host wiring

### 3. Configure the app

Magento 2 projects usually follow with:

```bash
govard config auto
```

### 4. Enter the workspace

```bash
govard shell
```

### 5. Open the app

Govard routes project domains through the shared proxy:

- app URL: `https://<project>.test`
- mail: `govard open mail`
- DB access: `govard open db`

## Daily Workflow

```bash
govard up
govard logs php -f
govard debug on
govard shell
govard down
```

Root shortcuts map to `govard env`:

- `govard up`
- `govard down`
- `govard restart`
- `govard ps`
- `govard logs`

## First Troubleshooting Checks

```bash
govard doctor
govard doctor trust
```

Use `govard doctor trust` if the browser does not trust local HTTPS yet.

## Next Docs

- [Commands](commands.md)
- [Configuration](configuration.md)
- [Frameworks](frameworks.md)
- [SSL and Domains](ssl-and-domains.md)
- [Desktop](desktop.md)
