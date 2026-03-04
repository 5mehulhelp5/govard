package tests

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"govard/internal/ui"
)

func TestPrintBrandIncludesTaglineAndVersion(t *testing.T) {
	version := "9.9.9-test"

	output, err := captureStdoutForTest(func() {
		ui.PrintBrand(version)
	})
	if err != nil {
		t.Fatalf("capture stdout: %v", err)
	}

	if !strings.Contains(output, "Go-based Versatile Runtime & Development") {
		t.Fatalf("expected tagline in banner output, got: %q", output)
	}
	if !strings.Contains(output, "____  ______     ___    ____  ____") {
		t.Fatalf("expected install-style ASCII logo in banner output, got: %q", output)
	}
	if !strings.Contains(output, "========================================") {
		t.Fatalf("expected separator in banner output, got: %q", output)
	}
	if !strings.Contains(output, "v"+version) {
		t.Fatalf("expected version in banner output, got: %q", output)
	}
}

func captureStdoutForTest(run func()) (string, error) {
	originalStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return "", err
	}
	defer readPipe.Close()

	os.Stdout = writePipe
	defer func() {
		os.Stdout = originalStdout
	}()

	run()

	if err := writePipe.Close(); err != nil {
		return "", err
	}

	var output bytes.Buffer
	if _, err := io.Copy(&output, readPipe); err != nil {
		return "", err
	}

	return output.String(), nil
}
