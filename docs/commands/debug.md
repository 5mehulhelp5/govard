# govard debug

Manage Xdebug settings and debugging sessions.

## Usage

```bash
govard debug on
govard debug off
govard debug status
govard debug shell
```

## Subcommands

### `on`

Enable Xdebug in the local environment. This updates `.govard.yml` and triggers `govard env up` to apply changes.

### `off`

Disable Xdebug in the local environment.

### `status`

Check whether Xdebug is currently enabled or disabled.

### `shell`

Open an interactive bash shell in the PHP container with Xdebug specific configurations active (targets the `<project>-php-debug-1` container if available).

## Notes

- Enabling/disabling Xdebug requires a container restart, which Govard handles automatically.
- Xdebug is typically used for step-debugging with IDEs like PHPStorm or VS Code.
