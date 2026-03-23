package tests

import (
	"strings"
	"testing"

	"govard/internal/engine"
)

func TestSanitizeSQLLineDefinerReplacement(t *testing.T) {
	line := "/*!50013 DEFINER=`alice`@`%` SQL SECURITY DEFINER */\n"
	sanitized, keep := engine.SanitizeSQLLine(line)
	if !keep {
		t.Fatal("expected line to be kept")
	}
	if !strings.Contains(sanitized, "/*!50013 */") {
		t.Fatalf("expected definer replacement, got: %s", sanitized)
	}
}

func TestSanitizeSQLLineDropsGTIDLines(t *testing.T) {
	line := "SET @@GLOBAL.GTID_PURGED='abc';\n"
	_, keep := engine.SanitizeSQLLine(line)
	if keep {
		t.Fatal("expected GTID line to be dropped")
	}
}

func TestSanitizeSQLDumpStream(t *testing.T) {
	input := strings.Join([]string{
		"CREATE TABLE test (id int);",
		"/*!50013 DEFINER=`alice`@`%` SQL SECURITY DEFINER */",
		"SET @@SESSION.SQL_LOG_BIN= 0;",
		"INSERT INTO test VALUES (1);",
		"",
	}, "\n")

	var output strings.Builder
	if err := engine.SanitizeSQLDump(strings.NewReader(input), &output); err != nil {
		t.Fatalf("sanitize dump: %v", err)
	}

	result := output.String()
	if strings.Contains(result, "@@SESSION.SQL_LOG_BIN") {
		t.Fatalf("expected SQL_LOG_BIN line removed, got: %s", result)
	}
	if !strings.Contains(result, "/*!50013 */") {
		t.Fatalf("expected definer replacement, got: %s", result)
	}
	if !strings.Contains(result, "INSERT INTO test VALUES (1);") {
		t.Fatalf("expected remaining payload preserved, got: %s", result)
	}
}
