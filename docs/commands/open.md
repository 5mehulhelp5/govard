# govard open

Open common service URLs in your browser.

## Usage

```bash
govard open admin
govard open db
govard open mail
govard open pma
govard open db -e local
govard open db -e staging
govard open shell -e dev
govard open sftp -e staging
govard open elasticsearch
govard open opensearch
```

## Targets

- `admin` Framework admin UI (https://<domain>/admin)
- `db` Database access (remote tunnel or local PHPMyAdmin)
- `mail` Local Mailpit UI (`https://mail.govard.test`)
- `pma` Local PHPMyAdmin target (`https://pma.govard.test`)
- `shell` Open local container shell or remote shell
- `sftp` Open remote SFTP URL in local app
- `elasticsearch` Elasticsearch endpoint
- `opensearch` OpenSearch endpoint

`target` is case-insensitive. Unknown targets return an error.

Environment behavior (`-e/--environment`):
- Omitted: local behavior for all targets.
- `-e local`: force local behavior.
- `-e <remote>`: use remote by name or by `remotes.<name>.environment` alias.

For `db`:
- Without `-e/--environment`: opens local PHPMyAdmin.
- With `-e local`: opens local PHPMyAdmin (`https://pma.govard.test`).
- With `-e <remote>`: starts an SSH tunnel first, then opens a `mysql://...` URL for local DB clients (for example BeeKeeper Studio). Keep the command running; `Ctrl+C` closes the tunnel.

For `pma`:
- Local only (`-e` omitted or `-e local`): opens `https://pma.govard.test`.
- Remote (`-e <remote>`): not supported. Use `govard open db -e <remote>`.

For `admin`:
- Local: opens `https://<domain>/admin`.
- Local Magento2: auto-detects backend path from `app/etc/env.php` (`backend.frontName`) and also checks DB config (`admin/url/custom*`) when available.
- Remote: opens admin URL from remote host (Magento2 also probes remote `frontName` when possible).
- Remote lookup supports remote name (`-e staging`) or environment alias (`-e prod`) when unique.

For `shell`:
- Local: opens local app container shell.
- Remote: opens SSH shell on remote target.

For `sftp`:
- Local: prints guidance that local SFTP target is not supported.
- Remote: opens `sftp://...` URL for configured remote.

For `elasticsearch`/`opensearch`:
- Local: opens local service URL.
- Remote: not supported yet.

## Examples

```bash
govard open admin
govard open db
govard open admin -e dev
govard open shell -e dev
govard open sftp -e staging
```
