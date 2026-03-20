# govard tunnel

Manage public tunnels and automatic base URL registration for local projects.

## Usage

```bash
govard tunnel start
govard tunnel status
govard tunnel stop
govard tunnel start --provider cloudflare --plan
```

## Subcommands

### `start [url]`

Starts a public tunnel and **automatically updates your local project base URL** to match the tunnel address. When the tunnel is stopped, the base URL is automatically reverted to the original value.

Defaults:
- Provider: `cloudflare`
- Target URL: `https://<domain>` from `.govard.yml`
- TLS verification: disabled (`--no-tls-verify=true`)

### `status`

Check if a tunnel is currently active and display the public URL.

### `stop`

Stop the active tunnel and revert the project base URL.

## Options

- `--provider` Tunnel provider (currently `cloudflare`)
- `--url` Target URL to expose (mutually exclusive with positional `[url]`)
- `--no-tls-verify` Disable TLS verification against the target URL
- `--plan` Print tunnel execution plan without running the provider command

## Notes

- **Automatic Base URL:** Govard detects your framework (Magento, Laravel, etc.) and updates the relevant configuration (e.g., `web/unsecure/base_url` in Magento) while the tunnel is active.
- Tunnels run as background processes. Use `tunnel stop` to clean up.
