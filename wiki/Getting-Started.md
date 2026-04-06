# Getting Started

This guide walks you through the shortest path from installation to a working local Govard project.

---

## 1. Initialize a Project

Navigate to your project root and run:

```bash
cd /path/to/your/project
govard init
```

Govard inspects `composer.json` or `package.json`, detects the framework, and writes `.govard.yml`.

### Detected Frameworks

| Framework | Detection |
| :--- | :--- |
| Magento 2 | `composer.json` with `magento/magento2-base` |
| Magento 1 / OpenMage | `composer.json` patterns |
| Laravel | `artisan` file + `composer.json` |
| Next.js | `package.json` with `next` dependency |
| Emdash | Emdash project markers |
| Drupal | `composer.json` with `drupal/core` |
| Symfony | `symfony/framework-bundle` |
| Shopware | `shopware/core` |
| CakePHP | `cakephp/cakephp` |
| WordPress | `wp-config.php` or `wp-login.php` |
| Custom | Interactive stack picker (`govard init --framework custom`) |

### Force a Specific Framework

```bash
govard init --framework magento2
govard init --framework laravel
govard init --framework custom
```

---

## 2. Start the Environment

```bash
govard env up
```

This renders a per-project compose file under `~/.govard/compose/` and starts your specialized stack.

### Common Variants

```bash
govard up --quickstart           # Alias: govard env up
govard env up --pull             # Pull latest images first
govard env up --fallback-local-build  # Build images locally if pull fails
```

### Startup Pipeline

1. Detect framework context
2. Validate config, Docker, ports, and prerequisites
3. Render compose file into `~/.govard/compose/`
4. Start containers in detached mode
5. Verify proxy and host wiring

### Root Shortcuts

| Shortcut | Equivalent |
| :--- | :--- |
| `govard up` | `govard env up` |
| `govard down` | `govard env down` |
| `govard restart` | `govard env restart` |
| `govard ps` | `govard env ps` |
| `govard logs` | `govard env logs` |

---

## 3. Configure the App

### Magento 2 Projects

Auto-inject container settings into `app/etc/env.php`:

```bash
govard config auto
```

### View Current Config

```bash
govard config get php_version
govard config get stack.db_type
```

---

## 4. Enter the Workspace

```bash
govard shell
```

- **PHP frameworks** (Magento, Laravel, etc.): opens the `php` container at `/var/www/html`
- **Node-first frameworks** (Next.js, Emdash): opens the `web` container at `/app`

---

## 5. Open Your App

Govard routes project domains through the shared Caddy proxy:

| Target | Command |
| :--- | :--- |
| App URL | `https://<project>.test` in browser |
| Admin panel | `govard open admin` |
| Mail (Mailpit) | `govard open mail` |
| Database (PHPMyAdmin) | `govard open db` |
| Database client URL | `govard open db --client` |

---

## 🔁 Daily Workflow

```bash
# Start work
govard up

# Follow logs
govard logs php -f

# Toggle Xdebug
govard debug on

# Enter shell
govard shell

# Stop work
govard down
```

---

## 🌐 Bootstrap a Remote Clone

To clone an existing environment from a remote server:

```bash
govard bootstrap --clone -e staging --no-pii --no-noise
```

For a fresh framework installation:

```bash
govard bootstrap --framework magento2 --fresh --framework-version 2.4.9
govard env up
govard open admin
```

---

## 🩺 First Troubleshooting Checks

```bash
govard doctor
govard doctor trust   # Fix browser SSL warnings
```

If your browser shows HTTPS trust warnings after setup, run `govard doctor trust` to re-import the Root CA.

---

## 📋 What's Next

| Topic | Link |
| :--- | :--- |
| All CLI commands | [CLI Commands](CLI-Commands) |
| Configuration options | [Configuration](Configuration) |
| Framework-specific notes | [Frameworks](Frameworks) |
| SSL and DNS setup | [SSL and Domains](SSL-and-Domains) |
| Remote environments | [Remotes and Sync](Remotes-and-Sync) |
| Desktop app | [Desktop App](Desktop-App) |

---

**[← Installation](Installation)** | **[CLI Commands →](CLI-Commands)**
