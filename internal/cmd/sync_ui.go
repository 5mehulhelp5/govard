package cmd

import (
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

type tailWriter struct {
	area    *pterm.AreaPrinter
	maxLine int
	lines   []string
	mu      sync.Mutex
	buffer  string
}

func newTailWriter(area *pterm.AreaPrinter, maxLine int) *tailWriter {
	return &tailWriter{
		area:    area,
		maxLine: maxLine,
		lines:   make([]string, 0, maxLine),
	}
}

func (w *tailWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer += string(p)

	// Keep processing as long as there's a delimiter (\n or \r)
	for {
		idxN := strings.Index(w.buffer, "\n")
		idxR := strings.Index(w.buffer, "\r")

		if idxN == -1 && idxR == -1 {
			break
		}

		var line string
		var useR bool
		if idxN != -1 && (idxR == -1 || idxN < idxR) {
			line = w.buffer[:idxN]
			w.buffer = w.buffer[idxN+1:]
		} else {
			line = w.buffer[:idxR]
			w.buffer = w.buffer[idxR+1:]
			useR = true
		}

		trimmedLine := strings.TrimRight(line, " \t")
		if trimmedLine == "" {
			continue
		}

		// Heuristic to handle rsync output and keep only relevant bits for UI
		// We want to skip some empty noise or purely numeric progress lines if possible,
		// but since the user specifically asked for 'follow last 10 lines', we'll be liberal.

		if useR && len(w.lines) > 0 {
			// \r typically means overwrite the last line (like progress)
			w.lines[len(w.lines)-1] = pterm.Gray("  > ") + trimmedLine
		} else {
			w.lines = append(w.lines, pterm.Gray("  > ")+trimmedLine)
			if len(w.lines) > w.maxLine {
				w.lines = w.lines[len(w.lines)-w.maxLine:]
			}
		}

		w.area.Update(strings.Join(w.lines, "\n"))
	}

	return len(p), nil
}
