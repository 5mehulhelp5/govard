# Govard Desktop

The desktop app is Wails-based and reuses the same core engine as the CLI.

Current status:

- Wails entrypoint in `cmd/govard-desktop`
- Dashboard UI in `desktop/frontend`
- `govard desktop` command to launch app (`--dev` for Wails dev mode)
- Optional background mode via `govard desktop --background` (start hidden, hide-on-close, single-instance reopen)
- Shell user preferences stored in `~/.govard/desktop-preferences.json`
- Desktop settings include theme, proxy target, preferred browser, and database client preference.

Keyboard shortcuts:

- `Ctrl+,` / `Cmd+,` opens Settings
- `Esc` closes Settings

Lightweight dashboard features:

- Environment list with start/stop/open
- Quick actions (Mailpit, PHPMyAdmin, DB Client, Xdebug toggle, health)
- Embedded Mailpit inbox panel with refresh/open controls
- Project onboarding panel (folder picker + add/init when `.govard.yml` is missing)
- Project onboarding supports optional Git clone (SSH/HTTPS URL + pre-clone connection validation) before add/init flow
- Git onboarding requires explicit folder-override confirmation and shows step progress (validate, prepare folder, clone, init)
- Remotes tab (list/add remotes, connectivity test, open SSH/DB/SFTP actions, and sync plan presets)
- Resource monitor (CPU/RAM/NET per project, OOM hints, refresh + auto-refresh)
- Logs with multi-service (`all`) selection, severity filtering, text search, and live streaming
- Shell launcher (service, user, shell)
- Native notifications from operation success/failure events while desktop is running
- Warnings panel
- Settings drawer
- Background update checks after startup with a bottom-right prompt for one-click download/install

Remote open behavior in Desktop:

- Open Database (Local) resolves Docker published host/port robustly and falls back to reachable container IPs when needed.
- Open Database (Local) falls back to PHPMyAdmin if launching the configured desktop DB client fails.
- Open Database (Remote) triggers `govard open db -e <remote> --client`.
- Open SSH (Remote) on Linux prefers native terminal launchers (`x-terminal-emulator`, `gnome-terminal`, `konsole`, `xfce4-terminal`) and falls back to `ssh://` URL handoff if needed.
- Open SFTP (Remote) prefers FileZilla when `filezilla` is available and falls back to `sftp://` URL handoff when not available.
- For remotes with `auth.method: ssh-agent`, desktop open actions reuse `SSH_AUTH_SOCK`.
- On Linux, when `SSH_AUTH_SOCK` is not set, desktop also probes `/run/user/<uid>/keyring/ssh`.
- If FileZilla still asks for a password, switch remote auth to `ssh-agent` and ensure your key is loaded into ssh-agent before launching Desktop.
- If no ssh-agent socket is available for `auth.method: ssh-agent`, desktop returns a clear error and you should load keys into ssh-agent or switch remote auth to `keychain`/`keyfile`.

Removed from desktop surface:

- Operations UI and operations output/history views
- Workflow snapshot cards
- Role-mode gated UI complexity

Frontend architecture:

- `desktop/frontend/main.js` bootstraps modules and event wiring
- `desktop/frontend/services/bridge.js` wraps Wails calls
- `desktop/frontend/state/store.js` holds UI state
- `desktop/frontend/modules/` contains feature modules
- `desktop/frontend/ui/toast.js` handles toasts
- `desktop/frontend/utils/dom.js` contains shared DOM helpers
