# Desktop App

Govard Desktop is the Wails-based GUI that reuses the same core engine as the CLI.

---

## Launch Modes

```bash
govard desktop              # Launch the built desktop binary
govard desktop --dev        # Run Wails dev mode (live backend)
govard desktop --background # Start hidden, reuse running instance on relaunch
```

| Mode | Description |
| :--- | :--- |
| `govard desktop` | Standard launch — uses built binary |
| `govard desktop --dev` | Dev mode — live Go backend, hot reload frontend |
| `govard desktop --background` | Background process — keeps alive when window closes |

---

## Current Surface

The desktop focuses on operational essentials:

| Feature | Description |
| :--- | :--- |
| **Environment Dashboard** | Start/stop/open/delete — including orphaned Docker project detection |
| **Project Workspace** | Environments list, quick actions, onboarding flow |
| **Quick Actions** | PHPMyAdmin, Xdebug toggle, health check, Mailpit, DB client |
| **Remotes Tab** | Add/test/open/sync-plan workflows for remote environments |
| **Resource Monitor** | CPU, RAM, network, OOM hints |
| **Logs** | Multi-service selection, severity filtering, text search, live streaming |
| **Shell Launcher** | Service, user, and shell selection |
| **Native Notifications** | Operation success/failure alerts |
| **Settings Drawer** | Theme, proxy target, preferred browser, database client |

> Desktop environment start/stop/pull and global-services operations call the Govard CLI command surface (`govard up`, `govard env ...`, `govard svc ...`), keeping desktop behavior aligned with all CLI updates.

---

## Keyboard Shortcuts

| Shortcut | Action |
| :--- | :--- |
| `Ctrl+,` / `Cmd+,` | Open Settings |
| `Esc` | Close Settings |

---

## Desktop Remote Actions

| Action | Behavior |
| :--- | :--- |
| Open Database (Remote) | Calls `govard open db -e <remote> --client` |
| Open SSH (Remote) | Prefers native Linux terminal launchers, falls back to `ssh://` |
| Open SFTP (Remote) | Prefers FileZilla, falls back to `sftp://` |

For `auth.method: ssh-agent`, Desktop reuses `SSH_AUTH_SOCK` and also probes `/run/user/<uid>/keyring/ssh` on Linux.

### Local Database Open

- Resolves published Docker host and port first
- Falls back to PHPMyAdmin if the configured DB client fails

---

## Desktop Preferences

Preferences are stored in:

```
~/.govard/desktop-preferences.json
```

Current persisted preferences:
- Theme (light/dark)
- Proxy target
- Preferred browser
- Database client preference

---

## Dev Mode

When developing the desktop app from source, provide a display server to prevent Wails from crashing in headless environments:

```bash
DISPLAY=:1 govard desktop --dev
```

Wails dev mode compiles the backend and exposes the frontend at:

```
http://localhost:34115
```

This is the preferred browser-testing path because the Go backend bridge stays live and loads real project data.

---

## Frontend Layout

| File | Purpose |
| :--- | :--- |
| `desktop/frontend/index.html` | Main HTML entry |
| `desktop/frontend/main.js` | Bootstrap, event wiring, tab/state management |
| `desktop/frontend/services/bridge.js` | Wails Go backend RPC bridge |
| `desktop/frontend/state/store.js` | Shared UI state (selected project, filters) |
| `desktop/frontend/modules/` | Feature modules (dashboard, logs, remotes, etc.) |
| `desktop/frontend/ui/toast.js` | Toast notification system |
| `desktop/frontend/utils/dom.js` | Shared DOM helpers |

### Test Mode Behavior

| Access Method | Backend | Data |
| :--- | :--- | :--- |
| Wails dev (`localhost:34115`) | Full backend bridge active | Real project data |
| Direct file (no backend) | Bridge unavailable | Mock fallback data + warning toast |

---

## Architecture Notes

The desktop app is intentionally focused on operational workflows:

- Desktop entrypoint: `cmd/govard-desktop`
- Wails bindings: `internal/desktop`
- Frontend shell: `desktop/frontend/index.html`
- Bootstrap/events: `desktop/frontend/main.js`
- Backend bridge: `desktop/frontend/services/bridge.js`
- State: `desktop/frontend/state/store.js`
- Feature modules: `desktop/frontend/modules/`

For deeper architecture context, see [Architecture](Architecture).

---

**[← SSL and Domains](SSL-and-Domains)** | **[Architecture →](Architecture)**
