# SSL and Domains

Govard provides local HTTPS for `.test` domains through the shared Caddy proxy and its internal certificate authority.

---

## What Govard Handles Automatically

- Local `.test` DNS routing via `dnsmasq`
- Certificate issuance for all project domains
- Root CA export to `~/.govard/ssl/root.crt`
- System trust-store installation (best-effort)
- Browser NSS import when `certutil` is available

---

## DNS Configuration for `.test` Domains

Govard runs a built-in `dnsmasq` service that resolves `*.test` domains to your local environment. You need to tell your OS to forward `.test` queries to this service.

### Linux — systemd-resolved (Recommended)

Works on Ubuntu, Debian, Arch, Fedora:

```bash
sudo mkdir -p /etc/systemd/resolved.conf.d
cat <<'EOF' | sudo tee /etc/systemd/resolved.conf.d/govard-test.conf
[Resolve]
DNS=127.0.0.1
Domains=~test
EOF
sudo systemctl restart systemd-resolved
```

### Linux — resolvconf (Legacy Ubuntu/Debian)

```bash
sudo apt-get install resolvconf
echo "nameserver 127.0.0.1" | sudo tee /etc/resolvconf/resolv.conf.d/tail
sudo resolvconf -u
```

### macOS

```bash
sudo mkdir -p /etc/resolver
echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/test
```

### Verify DNS Resolution

```bash
resolvectl query laravel.test
dig +short laravel.test
```

---

## Install Root CA Trust

`govard svc up` and `govard svc restart` auto-trust the Govard Root CA by default.

```bash
govard svc up         # Auto-trusts CA
govard doctor trust   # Manual trust (re-run anytime)
```

Skip auto-trust when needed:

```bash
govard svc up --no-trust
```

**What `doctor trust` does:**
1. Exports Root CA from Caddy to `~/.govard/ssl/root.crt`
2. Installs into system trust store (Linux/macOS)
3. Best-effort import into Chromium/Firefox NSS stores when `certutil` is available

> [!TIP]
> On Linux, install `certutil` from the `libnss3-tools` package so Govard can import into browser NSS stores automatically:
> ```bash
> sudo apt-get install libnss3-tools
> ```

---

## Browser Trust Configuration

If the OS trust is installed but your browser still shows warnings:

1. Locate **`~/.govard/ssl/root.crt`**
2. Open browser certificate settings (e.g., `chrome://settings/certificates`)
3. Navigate to the **Authorities** tab → click **Import**
4. Select `root.crt` and mark it trusted for websites
5. Restart the browser

Once trusted, all `*.test` domains managed by Govard will show a "Green Lock" without further configuration.

---

## Domain Management

### Extra Domains

```bash
govard domain add brand-b.test
govard domain remove brand-b.test
govard domain list
```

Govard routes these domains through the same proxy and CA flow as the primary project domain.

### Multi-Store Magento

For Magento multi-site setups:
- Use `store_domains` to automatically route hostnames and set scoped base URLs
- Use object entries (`type: website` or `type: store`) for automatic `MAGE_RUN_CODE` / `MAGE_RUN_TYPE` injection
- Use `extra_domains` only for additional hostnames **not** already in `store_domains`

```yaml
store_domains:
  brand-b.test:
    code: brand_b
    type: store
```

You do **not** need manual `SetEnvIf` rules in `.htaccess` for the standard typed `store_domains` flow.

---

## How Routing Works

1. `govard env up` renders the project stack and registers all routes
2. `govard env start` and `govard env restart` re-apply routes + local host entries after lifecycle changes
3. Caddy terminates HTTPS
4. Caddy forwards traffic to the project web container
5. Govard manages the local CA and exported root certificate

---

## Troubleshooting

### Browser says "Connection is not private"

Check in this order:

```bash
govard svc up               # Ensure global services are running
govard doctor trust         # Re-import Root CA
ls ~/.govard/ssl/root.crt   # Verify CA file exists
```

If still failing:
- Manually import `~/.govard/ssl/root.crt` into the browser
- Install `certutil` (Linux: `sudo apt-get install libnss3-tools`)
- Restart the browser

### Domain does not resolve

Check:
- `.test` resolver configuration (see [DNS Configuration](#dns-configuration-for-test-domains))
- `govard svc up` is running (includes the dnsmasq service)

```bash
govard svc up
resolvectl query myproject.test
```

### Certificate was not generated

```bash
govard env up
govard env logs
docker ps | grep caddy
```

### HTTPS not working after container restart

```bash
govard env restart    # Re-applies proxy routes + local domain entries
```

---

**[← Remotes and Sync](Remotes-and-Sync)** | **[Desktop App →](Desktop-App)**
