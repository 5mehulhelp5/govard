# govard deploy

Run deploy lifecycle hooks and static content deployment for the current project.

## Usage

```bash
govard deploy
govard deploy --locales "en_US fr_FR vi_VN"
govard deploy --strategy deployer
```

## Options

- `--strategy` Deployment strategy (`native` or `deployer`). Default: `native`.
- `--deployer` Shortcut flag for `--strategy deployer`.
- `--deployer-config` Path to Deployer config file.
- `-l, --locales` Space-separated locale codes for static content deployment (Magento 2). When not set, locales are **auto-detected from the local database**.

## Locale auto-detection (Magento 2)

When `--locales` is not specified and the framework is `magento2`, Govard automatically queries the local database container for active locale codes stored in `core_config_data`. The result always includes `en_US` as a baseline. If the database is unavailable, auto-detection is skipped silently.

```bash
# Auto-detects locales from local DB (e.g. "en_US fr_FR")
govard deploy

# Explicit locale override
govard deploy -l "en_US fr_FR vi_VN"
```

## Current behavior

- The command runs `pre_deploy` and `post_deploy` hooks.
- Strategy-related flags are accepted for compatibility and future extension; execution currently uses native flow.

## Examples

```bash
govard deploy
govard deploy -l "en_US"
govard deploy -l "en_US fr_FR vi_VN"
govard deploy --strategy deployer --deployer-config deploy.php
```
