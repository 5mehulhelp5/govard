# govard domain

Manage additional domains for the project.

## Usage

```bash
govard domain add <domain>
govard domain remove <domain>
govard domain list
```

## Subcommands

### `add <domain>`

Adds an extra domain to the project configuration.

```bash
govard domain add brand-b.test
```

- Writes the domain to `extra_domains` in `.govard.yml`.
- Automatically calls `govard env up` logic for the new domain (hosts mapping + proxy registration).

### `remove <domain>`

Removes an extra domain from the project configuration.

```bash
govard domain remove brand-b.test
```

- Removes the domain from `extra_domains` in `.govard.yml`.
- Unregisters the domain from the proxy and removes hosts entries.

### `list`

Lists all domains associated with the project, including the primary domain.

```bash
govard domain list
```

## Magento Multistore Support

For Magento 2 projects, you can map domains to specific store codes by manually editing `.govard.yml`:

```yaml
domain: main.test
extra_domains:
  - brand-b.test
store_domains:
  brand-b.test: brand_b_store
```

When `store_domains` is configured:

1. `govard config auto` will set the correct base URLs for each mapped store.
2. Web traffic for `brand-b.test` will be routed to the same environment.
