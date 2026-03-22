package engine

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

var (
	sqlDefinerPattern       = regexp.MustCompile(`DEFINER\s*=\s*[^*]+\*`)
	sqlSandboxPattern       = regexp.MustCompile(`.*999999.*sandbox.*`)
	sqlRowFormatPattern     = regexp.MustCompile(`ROW_FORMAT\s*=\s*FIXED`)
	sqlCollation0900Pattern = regexp.MustCompile(`utf8mb4_0900_ai_ci`)
	sqlCollation520Pattern  = regexp.MustCompile(`utf8(mb4)?_unicode_520_ci`)
)

func SanitizeSQLDump(input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			sanitized, keep := SanitizeSQLLine(line)
			if keep {
				if _, writeErr := io.WriteString(output, sanitized); writeErr != nil {
					return writeErr
				}
			}
		}

		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func SanitizeSQLLine(line string) (string, bool) {
	if strings.Contains(line, "@@GLOBAL.GTID_PURGED") || strings.Contains(line, "@@SESSION.SQL_LOG_BIN") {
		return "", false
	}
	if sqlSandboxPattern.MatchString(line) {
		return "", false
	}

	line = sqlDefinerPattern.ReplaceAllString(line, "*")
	line = sqlRowFormatPattern.ReplaceAllString(line, "")
	line = sqlCollation0900Pattern.ReplaceAllString(line, "utf8mb4_general_ci")
	line = sqlCollation520Pattern.ReplaceAllString(line, "utf8${1}_general_ci")

	return line, true
}
