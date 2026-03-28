# Desktop

Govard Desktop is the Wails-based GUI that reuses the same core engine as the CLI.

## Launch Modes

```bash
govard desktop
govard desktop --dev
govard desktop --background
```

- `govard desktop`: launch the built desktop binary
- `govard desktop --dev`: run Wails dev mode
- `govard desktop --background`: start hidden, keep the process alive when the window closes, and reuse the running instance on relaunch

## Current Surface

The lightweight desktop surface is intentionally focused on operational essentials:

- Environment dashboard with start/stop/open/delete
- Project workspace layout (environments, quick actions, onboarding)
- Quick actions (PHPMyAdmin, Xdebug toggle, health)
- Additional quick actions for Mailpit and DB client launch
- Remotes tab for add/test/open/sync-plan workflows
- Resource monitor with CPU, RAM, network, and OOM hints
- Logs with multi-service selection, severity filtering, text search, and live streaming
- Shell launcher (service, user, shell)
- Native notifications for operation success/failure updates
- Settings drawer for theme, proxy target, preferred browser, and database client preference

Desktop environment start/stop/pull and global-services start/stop/restart/pull now call the Govard CLI command surface instead of bypassing it with desktop-only compose shortcuts. This keeps desktop behavior aligned with the current `govard up`, `govard env ...`, and `govard svc ...` flows, including newer hook, proxy, trust, and fallback behavior added in CLI updates.

Desktop onboarding also supports an optional framework version field and forwards it to `govard init --framework-version ...` when provided.

Removed from the desktop surface:

- operations history UI
- workflow card sprawl
- role-gated UI complexity

## Keyboard Shortcuts

- `Ctrl+,` or `Cmd+,`: open Settings
- `Esc`: close Settings

## Desktop Remotes

Desktop remote actions call the same backend paths used by CLI commands.

- Open Database (Remote) uses `govard open db -e <remote> --client`
- Open SSH (Remote) on Linux prefers native terminal launchers and falls back to `ssh://`
- Open SFTP (Remote) prefers FileZilla and falls back to `sftp://`

For `auth.method: ssh-agent`, Desktop reuses `SSH_AUTH_SOCK` and also probes `/run/user/<uid>/keyring/ssh` on Linux.

## Local Database Open Behavior

- Open Database (Local) resolves published Docker host and port first
- if the configured desktop DB client fails, Govard falls back to PHPMyAdmin

## Dev Mode

When developing the desktop app from source:

```bash
DISPLAY=:1 govard desktop --dev
```

Wails dev mode exposes the frontend at:

```text
http://localhost:34115
```

This is the preferred browser-testing path for the real desktop UI because the Go backend bridge stays live and loads real project data.

## Frontend Layout

- `desktop/frontend/index.html`: main HTML entry
- `desktop/frontend/main.js`: bootstrap and event wiring
- `desktop/frontend/services/bridge.js`: Wails backend bridge
- `desktop/frontend/state/store.js`: shared UI state
- `desktop/frontend/modules/`: feature modules
- `desktop/frontend/ui/toast.js`: toast notifications
- `desktop/frontend/utils/dom.js`: DOM helpers

## Preferences

Desktop preferences are stored in:

```text
~/.govard/desktop-preferences.json
```

Current persisted preferences include theme, proxy target, preferred browser, database client preference, and shell user preferences.

## Related Docs

- [Commands](commands.md)
- [Remotes and Sync](remotes-and-sync.md)
- [Architecture](architecture.md)
