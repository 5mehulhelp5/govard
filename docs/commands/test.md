# govard test

Run project testing frameworks within the local environment.

## Usage

```bash
govard test phpunit
govard test phpstan
govard test mftf
govard test integration
```

## Subcommands

### `phpunit`

Run PHPUnit tests. Passes all additional arguments directly to the `phpunit` binary.

```bash
# Run all tests
govard test phpunit

# Run a specific test suite
govard test phpunit --testsuite unit
```

### `phpstan`

Run PHPStan static analysis.

### `mftf`

Run Magento Functional Testing Framework (MFTF) tests.

### `integration`

Run Magento integration tests.

## Notes

- All tests are executed within the project's PHP container.
- Govard automatically resolves the correct execution user (e.g., `www-data`) and working directory.
- For MFTF, Ensure Selenium/Allure services are enabled in your `.govard.yml`.
