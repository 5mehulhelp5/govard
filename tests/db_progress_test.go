package tests

import (
	"bytes"
	"io"
	"testing"

	"govard/internal/engine"

	"github.com/pterm/pterm"
)

func TestProgressReader(t *testing.T) {
	data := []byte("hello world")
	buf := bytes.NewReader(data)

	bar, _ := pterm.DefaultProgressbar.WithTotal(len(data)).Start()
	pr := engine.NewProgressReader(buf, bar)

	out := make([]byte, 5)
	n, err := pr.Read(out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes, got %d", n)
	}
	if bar.Current != 5 {
		t.Fatalf("expected bar current 5, got %d", bar.Current)
	}

	_, _ = io.ReadAll(pr)
	if bar.Current != len(data) {
		t.Fatalf("expected bar current %d, got %d", len(data), bar.Current)
	}
}
