# SSL and Domains

Govard provides local HTTPS for `.test` domains through the shared Caddy proxy and its internal certificate authority.

## What Govard Handles

- local `.test` DNS routing through `dnsmasq`
- certificate issuance for local project domains
- root CA export to `~/.govard/ssl/root.crt`
- best-effort system trust-store installation
- best-effort browser NSS import when `certutil` is available

## DNS for `.test`

Govard expects `.test` queries to resolve to the local loopback DNS service.

### Linux with `systemd-resolved`

```bash
sudo mkdir -p /etc/systemd/resolved.conf.d
cat <<'EOF' | sudo tee /etc/systemd/resolved.conf.d/govard-test.conf
[Resolve]
DNS=127.0.0.1
Domains=~test
EOF
sudo systemctl restart systemd-resolved
```

### macOS

```bash
sudo mkdir -p /etc/resolver
echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/test
```

### Verify DNS

```bash
resolvectl query laravel.test
dig +short laravel.test
```

## Install Trust

`govard svc up` and `govard svc restart` auto-trust the Govard Root CA unless you opt out.

```bash
govard svc up
govard doctor trust
```

Skip auto-trust when needed:

```bash
govard svc up --no-trust
```

## Browser Trust

If the OS trust is installed but the browser still warns:

1. Locate `~/.govard/ssl/root.crt`
2. Import it into the browser authority store
3. Mark it trusted for websites
4. Restart the browser

On Linux, install `certutil` from `libnss3-tools` so Govard can import into NSS stores automatically.

## Domain Management

Add extra project domains:

```bash
govard domain add brand-b.test
govard domain remove brand-b.test
govard domain list
```

Govard maps these domains through the same proxy and CA flow as the primary project domain.

## How Routing Works

1. `govard env up` renders the project stack and registers routes
2. Caddy terminates HTTPS
3. Caddy forwards traffic to the project web container
4. Govard manages the local CA and exported root certificate

## Troubleshooting

### Browser says the connection is not private

Check in this order:

```bash
govard svc up
govard doctor trust
ls ~/.govard/ssl/root.crt
```

If needed:

- import `~/.govard/ssl/root.crt` manually
- install `certutil`
- restart the browser

### Domain does not resolve

Check:

- the `.test` resolver config
- the local DNS service on `127.0.0.1`
- `govard svc up`

### Certificate was not generated

Check:

```bash
govard env up
govard env logs
docker ps | grep caddy
```

## Related Docs

- [Getting Started](getting-started.md)
- [Commands](commands.md)
