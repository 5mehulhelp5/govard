# govard debug

Manage Xdebug settings and debugging sessions. 

## Usage

When run without any subcommands, `govard debug` defaults to opening an interactive debug shell.

```bash
govard debug            # Open debug shell (shorthand)
govard debug on         # Enable Xdebug
govard debug off        # Disable Xdebug
govard debug status     # Check status
govard debug shell      # Open debug shell
```

## Subcommands

### `on`

Enable Xdebug in the local environment. This updates `.govard.yml` and triggers `govard env up` to apply changes.

### `off`

Disable Xdebug in the local environment.

### `status`

Check whether Xdebug is currently enabled or disabled.

### `shell`

Open an interactive bash shell in the PHP container with Xdebug specific configurations active (targets the `<project>-php-debug-1` container).

## IDE Configuration

Govard automatically configures the environment for seamless IDE integration (PHPStorm, VS Code).

### Server Name

The `PHP_IDE_CONFIG` environment variable is set to `serverName=<project_name>-docker`. In your IDE:
- Create a Server named exactly your **Project Name** plus `-docker` (e.g., `magento2-docker`).
- You can find the exact name for your project by running `govard debug status`.
- Set the Host to your project domain (e.g., `magento2.test`).
- Map the project root to `/var/www/html`.

### Xdebug Versions

Govard automatically detects the PHP version and applies the appropriate Xdebug settings:
- **PHP >= 7.2**: Uses Xdebug 3 configuration (`XDEBUG_MODE`).
- **PHP < 7.2**: Uses Xdebug 2 configuration (`remote_enable`).

### Port Configuration

- **Port**: Always configure your IDE to listen on port **9003**.
- Govard maps both Xdebug 2 and Xdebug 3 to port `9003` to maintain consistency and avoid conflicts with the default PHP-FPM port (`9000`).

## Notes

- Enabling/disabling Xdebug requires a container restart, which Govard handles automatically.
- For Web debugging, ensure you use a browser extension (like Xdebug Helper) to set the `XDEBUG_SESSION=PHPSTORM` cookie.
