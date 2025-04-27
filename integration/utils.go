package integration

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// captureOutput captures the stdout output of the function f and returns it as a string.
// It ensures proper pipe handling and error checking.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()

	// Save the original stdout
	old := os.Stdout
	defer func() { os.Stdout = old }()

	// Create a pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Run the function
	f()

	// Close the write end of the pipe
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe writer: %v", err)
	}

	// Read the output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy pipe output: %v", err)
	}

	// Close the read end of the pipe
	if err := r.Close(); err != nil {
		t.Fatalf("Failed to close pipe reader: %v", err)
	}

	return buf.String()
}
