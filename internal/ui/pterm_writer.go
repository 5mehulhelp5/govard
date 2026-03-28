package ui

import (
	"strings"

	"github.com/pterm/pterm"
)

type ptermPrinter interface {
	Println(args ...any) *pterm.TextPrinter
}

type ptermWriter struct {
	printer ptermPrinter
}

func (w *ptermWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			w.printer.Println(trimmed)
		}
	}
	return len(p), nil
}

// NewPtermWriter creates an io.Writer that pipes every line to a pterm printer.
func NewPtermWriter(printer ptermPrinter) *ptermWriter {
	return &ptermWriter{printer: printer}
}
