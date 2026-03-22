package engine

import (
	"testing"
)

func TestSanitizeSQLLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
		keep     bool
	}{
		{
			name:     "Leave normal line",
			line:     "CREATE TABLE `test` (id int);",
			expected: "CREATE TABLE `test` (id int);",
			keep:     true,
		},
		{
			name:     "Remove GTID_PURGED",
			line:     "SET @@GLOBAL.GTID_PURGED='...';",
			expected: "",
			keep:     false,
		},
		{
			name:     "Remove SQL_LOG_BIN",
			line:     "SET @@SESSION.SQL_LOG_BIN= 0;",
			expected: "",
			keep:     false,
		},
		{
			name:     "Remove Sandbox lines",
			line:     "/*... 999999.*enable the sandbox mode ...*/",
			expected: "",
			keep:     false,
		},
		{
			name:     "Strip DEFINER",
			line:     "/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`localhost`*/ /*!50003 TRIGGER `test`*/",
			expected: "/*!50003 CREATE*/ /*!50017 */ /*!50003 TRIGGER `test`*/",
			keep:     true,
		},
		{
			name:     "Remove ROW_FORMAT=FIXED",
			line:     "CREATE TABLE `test` (id int) ENGINE=InnoDB ROW_FORMAT=FIXED;",
			expected: "CREATE TABLE `test` (id int) ENGINE=InnoDB ;",
			keep:     true,
		},
		{
			name:     "Map utf8mb4_0900_ai_ci",
			line:     "COLLATE=utf8mb4_0900_ai_ci",
			expected: "COLLATE=utf8mb4_general_ci",
			keep:     true,
		},
		{
			name:     "Map utf8mb4_unicode_520_ci",
			line:     "CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci",
			expected: "CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
			keep:     true,
		},
		{
			name:     "Map utf8_unicode_520_ci",
			line:     "COLLATE=utf8_unicode_520_ci",
			expected: "COLLATE=utf8_general_ci",
			keep:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, keep := SanitizeSQLLine(tt.line)
			if keep != tt.keep {
				t.Errorf("SanitizeSQLLine() keep = %v, want %v", keep, tt.keep)
			}
			if got != tt.expected {
				t.Errorf("SanitizeSQLLine() got = %v, want %v", got, tt.expected)
			}
		})
	}
}
