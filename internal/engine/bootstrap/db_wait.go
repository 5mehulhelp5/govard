package bootstrap

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"govard/internal/conventions"
)

// waitForMySQLDatabase polls a MySQL/MariaDB connection until it succeeds,
// retrying for up to 30 seconds. Used by frameworks whose fresh-install runs
// an installer against the DB before the DB container has finished starting.
func waitForMySQLDatabase(projectDir string, runner func(command string) error, host, user, pass, name string) error {
	code := strings.Join([]string{
		"mysqli_report(MYSQLI_REPORT_OFF);",
		"$db = mysqli_init();",
		"if (!$db) { exit(1); }",
		"if (!@mysqli_real_connect($db, " + strconv.Quote(host) + ", " + strconv.Quote(user) + ", " + strconv.Quote(pass) + ", " + strconv.Quote(name) + ", " + strconv.Itoa(conventions.MySQLPort) + ")) {",
		"    exit(1);",
		"}",
	}, "\n")

	var lastErr error
	for range 30 {
		if err := runPHPOneLiner(projectDir, runner, code); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(time.Second)
	}

	return fmt.Errorf("wait for database: %w", lastErr)
}
