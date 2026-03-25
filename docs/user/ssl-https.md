# 🔒 SSL & HTTPS

Govard provides automated local HTTPS for all `.test` domains using a built-in certificate authority (Caddy).

## 🚀 Key Features

- **Zero-Config**: Certificates are automatically generated for any `.test` domain.
- **Wildcard Support**: One CA trusts all your local projects.
- **Persistent Trust**: Once configured, new projects work instantly.

## 🛠️ Setup

### 1. DNS Resolver for `.test` Domains

Govard now runs a built-in `dnsmasq` service on the local loopback interface (port 53) to automatically resolve `*.test` domains to your local environment.

You need to configure your operating system to forward `.test` queries to this local service.

**Linux (Ubuntu/Debian with systemd-resolved - Recommended):**

```bash
sudo mkdir -p /etc/systemd/resolved.conf.d
cat <<'EOF' | sudo tee /etc/systemd/resolved.conf.d/govard-test.conf
[Resolve]
DNS=127.0.0.1
Domains=~test
EOF
sudo systemctl restart systemd-resolved
```

**Ubuntu (resolvconf - Legacy):**

```bash
sudo apt-get install resolvconf
echo "nameserver 127.0.0.1" | sudo tee /etc/resolvconf/resolv.conf.d/tail
sudo resolvconf -u
```

**Arch Linux (systemd-resolved):**

```bash
sudo systemctl enable --now systemd-resolved
sudo ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
sudo mkdir -p /etc/systemd/resolved.conf.d
cat <<'EOF' | sudo tee /etc/systemd/resolved.conf.d/govard-test.conf
[Resolve]
DNS=127.0.0.1
Domains=~test
EOF
sudo systemctl restart systemd-resolved
```

**Fedora (systemd-resolved):**

```bash
sudo systemctl enable --now systemd-resolved
sudo ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
sudo mkdir -p /etc/systemd/resolved.conf.d
cat <<'EOF' | sudo tee /etc/systemd/resolved.conf.d/govard-test.conf
[Resolve]
DNS=127.0.0.1
Domains=~test
EOF
sudo systemctl restart systemd-resolved
```

Verify DNS:

```bash
resolvectl query laravel.test
dig +short laravel.test
```

macOS (Create a resolver file):

```bash
sudo mkdir -p /etc/resolver
echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/test
```

### 2. Install the Root CA

By default, `govard svc up` and `govard svc restart` now auto-trust the Govard Root CA:

```bash
govard svc up
```

You can also run trust manually at any time:

```bash
govard doctor trust
```

What happens automatically:

- Exports Root CA from Caddy to `~/.govard/ssl/root.crt`
- Installs it into the system trust store (Linux/macOS)
- Best-effort import to browser NSS stores (Chromium/Firefox profiles) when `certutil` is available

Optional flags on `svc up`/`svc restart`:

```bash
govard svc up --no-trust
```

### 3. Browser Configuration

Govard now tries to import browser trust automatically. If your browser still shows trust warnings:

1. **Locate the CA**: `~/.govard/ssl/root.crt`.
2. **Open Settings**: Go to `chrome://settings/certificates` in your browser.
3. **Import**: Navigate to the **Authorities** tab and click **Import**.
4. **Select File**: Select `~/.govard/ssl/root.crt`.
5. **Trust**: Check **"Trust this certificate for identifying websites"**.
6. **Restart**: Restart your browser.

Tip for Linux: install `certutil` (package `libnss3-tools`) so Govard can auto-import into NSS stores.

## 🔍 Technical Details

Govard uses Caddy's internal PKI to manage local certificates. The root certificate is generated inside the global proxy Caddy container and exported to `~/.govard/ssl/root.crt`.

### How It Works

1. **Certificate Generation**: When you run `govard env up`, Caddy generates a certificate for your `.test` domain.
2. **Proxy Routing**: The global Caddy proxy routes `*.test` domains to your project containers.
3. **HTTPS Termination**: SSL/TLS is terminated at the Caddy proxy level.
4. **Internal Routing**: Caddy forwards requests to your project's web container.

### Certificate Location

Linux:

```
~/.govard/
└── ssl/
    └── root.crt    # Root CA certificate
```

### Troubleshooting

**Issue**: Browser shows "Your connection is not private"

**Solutions**:
1. Ensure proxy is running: `govard svc up`
2. Run `govard doctor trust` again
3. Install `certutil` (`libnss3-tools`) to enable browser auto-import
4. Manually import the certificate in your browser (see step 3 above)
5. Check that `~/.govard/ssl/root.crt` exists after `govard doctor trust`

**Issue**: Certificate not generated

**Solution**:
1. Ensure `govard env up` completed successfully
2. Check Caddy proxy is running: `docker ps | grep caddy`
3. Check logs: `govard env logs`
