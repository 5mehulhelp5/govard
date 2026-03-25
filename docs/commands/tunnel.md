# govard tunnel

Manage public tunnels and framework-specific base URL registration for local projects.

## Usage

```bash
govard tunnel start
govard tunnel status
govard tunnel stop
govard tunnel start --provider cloudflare --plan
```

## Subcommands

### `start [url]`

Starts a public tunnel. For supported frameworks, Govard also updates the local project base URL to match the tunnel address and reverts it when the tunnel stops.

Defaults:
- Provider: `cloudflare`
- Target URL: `https://<domain>` from `.govard.yml`
- TLS verification: disabled (`--no-tls-verify=true`)

### `status`

Check if a tunnel is currently active and display the public URL.

### `stop`

Stop the active tunnel and revert the project base URL for supported frameworks.

## Options

- `--provider` Tunnel provider (currently `cloudflare`)
- `--url` Target URL to expose (mutually exclusive with positional `[url]`)
- `--no-tls-verify` Disable TLS verification against the target URL
- `--plan` Print tunnel execution plan without running the provider command

## Notes

- **Automatic Base URL:** Govard currently mutates application base URL only for `magento1`, `magento2`, `laravel`, `symfony`, and `wordpress`.
- For other frameworks, Govard still starts the tunnel but does not change application config.
- Tunnels run as background processes. Use `tunnel stop` to clean up.
