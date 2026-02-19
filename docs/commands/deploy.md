# govard deploy

Run deploy lifecycle hooks for the current project.

## Usage

```bash
govard deploy
govard deploy --strategy native
govard deploy --strategy deployer
```

## Options

- `--strategy` Deployment strategy flag (`native` or `deployer`)
- `--deployer` Shortcut flag for `--strategy deployer`
- `--deployer-config` Deployer config path flag

Current behavior:
- The command runs `pre_deploy` and `post_deploy` hooks.
- Execution currently uses native flow regardless of strategy flags.
- Strategy-related flags are accepted for compatibility and future extension.

## Examples

```bash
govard deploy --strategy native
govard deploy --strategy deployer
```
