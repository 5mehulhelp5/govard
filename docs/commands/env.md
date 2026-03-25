# govard env (alias: project)

Manage the local development environment for the current project. Govard provides a smart wrapper around Docker Compose, handling framework-specific hooks, proxy registration, and host mapping.

Common root shortcuts are also available:

- `govard up` → `govard env up`
- `govard down` → `govard env down`
- `govard restart` → `govard env restart`
- `govard ps` → `govard env ps`
- `govard logs` → `govard env logs`

## Root Commands

### `up`

Starts all Docker containers required for the current project.

```bash
govard env up [flags]
govard up [flags]
```

**Stages:**
1. **Detect**: Identifies framework context (Magento, Laravel, etc.).
2. **Validate**: Checks Docker status, port conflicts, and configuration.
3. **Render**: Generates a specialized Docker Compose file in `~/.govard/compose/`.
4. **Start**: Runs `docker compose up -d` with Govard-managed settings.
5. **Verify**: Maps `.test` domains and registers them with the Govard Proxy.

**Options:**

- `--pull`: Pull latest images before starting.
- `--force-recreate`: Recreate containers even if their configuration and image haven't changed.
- `--quickstart`: Skip heavy services (Elasticsearch, Varnish, Redis) for faster startup.
- `--remove-orphans`: Remove containers for services no longer in config.

---

### `start`, `stop`, `restart`

Control the running state of the project without tearing down containers.

```bash
govard env start    # Start stopped containers
govard env stop     # Stop running containers
govard env restart  # Restart all project containers
```

---

### `down`

Tear down the current project environment (containers and networks).

```bash
govard env down [options]
```

**Options:**
- `-v, --volumes`: Remove named and anonymous volumes.
- `--rmi local`: Remove service images built locally.

---

### `logs`

Stream container logs with optional filtering.

```bash
govard env logs [service] [-f]
govard env logs --errors        # Filter for error/critical messages
```

---

### `ps`, `pull`, `build`

Standard maintenance commands proxied directly to Docker Compose.

```bash
govard env ps      # List project containers
govard env pull    # Pull latest images
govard env build   # Rebuild project images
```

## Service Shortcuts

Govard provides direct shortcuts for common services. These shortcuts intelligently proxy maintenance commands to Docker Compose while offering custom utility subcommands.

### `redis` / `valkey`

Manage the cache service.

```bash
govard redis cli      # Open interactive CLI
govard redis flush    # Flush all keys
govard redis logs -f  # View redis logs (proxy)
```

### `varnish`

Manage the Varnish edge cache.

```bash
govard varnish log    # View varnish logs
govard varnish ban    # Ban a URL pattern
govard varnish stats  # View varnish performance stats
```

### `elasticsearch` / `opensearch`

Interact with the search service.

```bash
govard elasticsearch _cat/indices  # Run a curl query
govard elasticsearch ps            # Check service status (proxy)
```

## Notes

- All `env` commands are executed from within the project directory.
- Govard manages the `.govard.yml` configuration and maps it to Docker Compose logic.
- Shortcuts like `govard redis` are available at the root level for convenience.
