package engine

import (
	"io"

	"github.com/pterm/pterm"
)

// ProgressReader wraps an io.Reader and updates a pterm progress bar as bytes are read.
type ProgressReader struct {
	Reader io.Reader
	Bar    *pterm.ProgressbarPrinter
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 && pr.Bar != nil {
		pr.Bar.Add(n)
	}
	return n, err
}

// NewProgressReader creates a new ProgressReader.
func NewProgressReader(r io.Reader, bar *pterm.ProgressbarPrinter) *ProgressReader {
	return &ProgressReader{
		Reader: r,
		Bar:    bar,
	}
}
