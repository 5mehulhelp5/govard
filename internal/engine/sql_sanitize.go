package engine

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

var sqlDefinerPattern = regexp.MustCompile("DEFINER[ ]*=[ ]*`[^`]+`@`[^`]+`")

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
	return sqlDefinerPattern.ReplaceAllString(line, "DEFINER=CURRENT_USER"), true
}
