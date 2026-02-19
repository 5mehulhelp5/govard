# govard down

Tear down the current project environment (containers and networks).

## Usage

```bash
govard down
govard down --volumes
govard down --rmi local --timeout 20
```

## Options

- `--remove-orphans` Remove containers for services no longer defined in compose (default: `true`)
- `-v, --volumes` Remove named and anonymous volumes
- `--rmi` Remove service images (`all` or `local`)
- `-t, --timeout` Shutdown timeout in seconds

## Notes

- `govard down` uses `docker compose down` with the project compose file under `~/.govard/compose/`.
- It also unregisters the project domain from Govard proxy and removes the hosts entry when possible.
- Lifecycle hooks `pre_stop` / `post_stop` are executed for consistency with `govard stop`.

## Related

- `govard stop` only stops containers without removing networks/containers.
- `govard up` recreates and starts the environment after teardown.
