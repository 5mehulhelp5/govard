package conventions

const (
	// MySQLDumpBinDetect detects mariadb-dump or mysqldump
	MySQLDumpBinDetect = `if command -v mariadb-dump >/dev/null 2>&1; then DUMP_BIN=mariadb-dump; else DUMP_BIN=mysqldump; fi`

	// MySQLClientBinDetect detects mariadb or mysql
	MySQLClientBinDetect = `if command -v mysql >/dev/null 2>&1; then DB_CLI=mysql; elif command -v mariadb >/dev/null 2>&1; then DB_CLI=mariadb; else echo "mysql client not found (mysql/mariadb)" >&2; exit 127; fi`
)
