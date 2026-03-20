# govard redis

Manage Redis or Valkey cache for local and remote environments.

## Usage

```bash
govard redis cli
govard redis flush
govard redis info
```

## Subcommands

### `cli`

Open an interactive CLI session to the Redis/Valkey service.

- Local: connects to the `<project>-redis-1` container.
- Remote: connects via SSH using the remote's `redis-cli` or `valkey-cli`.

### `flush`

Flush all keys from the cache.

- Supports both local and remote environments.
- Automatically handles both Redis and Valkey providers.

### `info`

Display information and statistics about the Redis/Valkey service.

## Options

- `-e, --environment` Target environment (default: local)

## Notes

- Govard automatically detects whether the service is Redis or Valkey and uses the appropriate CLI tool (`redis-cli` or `valkey-cli`).
- Remote operations require the `cache` capability to be enabled for the target environment.
