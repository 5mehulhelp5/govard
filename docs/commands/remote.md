# govard remote

Manage remote environments for sync, deploy, and database operations.

## Usage

```bash
govard remote add <name> --host <host> --user <user> --path <path>
govard remote exec <name> -- <command>
govard remote test <name>
govard remote audit tail --lines 20
govard remote audit stats --lines 200
govard remote audit tail --since 2026-02-12T00:00:00Z --until 2026-02-12T23:59:59Z
```

## Options

- `--host` Remote host (required)
- `--user` Remote user (required)
- `--path` Remote project root (required)
- `--port` Remote SSH port (default: 22)
- `--capabilities` Allowed remote scopes (`files,media,db,deploy` or `all`)
- `--auth-method` Remote auth method (`keychain`, `ssh-agent`, `keyfile`)
- `--key-path` SSH private key path (stored in auth store when using `--auth-method keychain`)
- `--strict-host-key` Enable strict SSH host key verification (`StrictHostKeyChecking=yes`)
- `--known-hosts-file` Custom SSH `known_hosts` file (implies `--strict-host-key`)
- `--protected` Prevents destructive writes to this remote (defaults to true for 'prod' remotes)

## Examples

```bash
govard remote add staging --host staging.example.com --user deploy --path /var/www/html
govard remote add staging --host staging.example.com --user deploy --path /var/www/html --strict-host-key --known-hosts-file ~/.ssh/known_hosts
govard remote add staging --host staging.example.com --user deploy --path /var/www/html --auth-method keychain --key-path ~/.ssh/id_ed25519
govard remote add ci --host ci.example.com --user deploy --path /srv/www/app --auth-method keyfile
govard remote add prod --host prod.example.com --user deploy --path /srv/www/app --capabilities files,media --protected=false
govard remote test staging
govard remote exec staging -- ls -la
govard remote audit tail --status failure --lines 50
govard remote audit stats --status failure --json
govard remote audit tail --since 2026-02-12 --until 2026-02-12
```

## Notes

- `remote test` validates SSH connectivity and checks remote `rsync` availability.
- `remote test` reports probe latency and classifies failures (`network`, `auth`, `permission`, `host_key`, `dependency`) with remediation hints.
- `remote exec` forwards your local SSH agent.
- SSH key path resolution priority is: `remotes.<name>.auth.key_path` -> `GOVARD_REMOTE_KEY_PATH_<REMOTE_NAME>` -> `GOVARD_REMOTE_KEY_PATH` -> keychain/file auth store (for `keychain`) -> default keyfile probe (`~/.ssh/id_ed25519`, `~/.ssh/id_ecdsa`, `~/.ssh/id_rsa`) for `keyfile`.
- Remote fields `host`, `user`, `path`, `auth.key_path`, `auth.known_hosts_file`, and `paths.media` support `op://...` references resolved via the 1Password CLI (`op read`).
- Production remotes (normalizing to `prod` via name) are write-protected by default. This can be overridden via `protected: false`.
- Remote operations are appended to `~/.govard/remote.log` (override with `GOVARD_REMOTE_AUDIT_LOG_PATH`).
- Remote commands also emit structured operation events to `~/.govard/operations.log`, which the desktop app uses for native success/failure notifications.
- In Desktop Remotes tab, `Open Database` uses `govard open db -e <remote> --client`.
- In Desktop Remotes tab, `Open SSH` on Linux prefers native terminal launchers and falls back to `ssh://` URL handoff.
- In Desktop Remotes tab, `Open SFTP` prefers FileZilla when available and falls back to `sftp://` URL handoff.
- For desktop remote open actions with `auth.method: ssh-agent`, Govard reuses `SSH_AUTH_SOCK` and also probes `/run/user/<uid>/keyring/ssh` on Linux.
- If FileZilla prompts for password unexpectedly, prefer `auth.method: ssh-agent` and confirm the key is loaded in your local ssh-agent.
- `remote audit tail` supports filtering by `--status` and `--operation`, and `--json` output.
- `remote audit stats` summarizes recent events by status, category, and operation.
- Both `tail` and `stats` support time-window filters via `--since` and `--until` (RFC3339 or `YYYY-MM-DD`).
