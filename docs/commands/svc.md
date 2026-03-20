# govard svc

Manage global shared services and workspace-wide state. Global services (Proxy, Mailpit, PHPMyAdmin, Portainer) are shared across all Govard projects.

## Global Service Commands

### `up`, `down`, `restart`

Manage the shared service stack.

```bash
govard svc up [flags]
govard svc down
govard svc restart
```

**`up` Options:**
- `--pull`: Pull latest global service images.
- `--no-trust`: Skip root CA trust installation.
- `--remove-orphans`: Remove obsolete global containers.

---

### `ps`, `logs`, `pull`

Standard maintenance commands for global services.

```bash
govard svc ps       # List global containers
govard svc logs -f  # Stream global service logs
govard svc pull     # Update global images
```

## Workspace State Management

### `sleep`

Stops all currently running Govard projects to save system resources. The "wake state" is persisted to `~/.govard/sleep-state.json`.

```bash
govard svc sleep
```

### `wake`

Restores all projects that were active when `sleep` was called.

```bash
govard svc wake
```

## Notes

- Global services are managed via a central Docker Compose stack in `~/.govard/proxy/`.
