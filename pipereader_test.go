package cronlogger_test

import (
	"cronlogger"
	"io"
	"os"
	"testing"
)

func Test_ReadPipeStdin_Success(t *testing.T) {
	// Save the original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Write test data to the pipe
	testData := "test input data"
	go func() {
		defer w.Close()
		// Write data in a single write operation
		w.Write([]byte(testData))
	}()

	// Call the function
	result, err := cronlogger.ReadStdin()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if result != testData {
		t.Errorf("expected %q, got %q", testData, result)
	}
}

func Test_ReadPipeStdin_EmptyInput(t *testing.T) {
	// Save the original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Close the write end immediately to simulate empty input
	w.Close()

	// Call the function
	result, err := cronlogger.ReadStdin()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func Test_ReadPipeStdin_LargeInput(t *testing.T) {
	// Save the original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Write large test data to the pipe (larger than buffer size)
	testData := make([]byte, 8*1024) // 8KB, larger than the 4KB buffer
	for i := range testData {
		testData[i] = byte('A' + (i % 26))
	}

	go func() {
		defer w.Close()
		w.Write(testData)
	}()

	// Call the function
	result, err := cronlogger.ReadStdin()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(result) != len(testData) {
		t.Errorf("expected length %d, got %d", len(testData), len(result))
	}

	if result != string(testData) {
		t.Error("result does not match expected data")
	}
}

func Test_ReadPipeStdin_MultilineInput(t *testing.T) {
	// Save the original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Write multiline test data
	testData := "line 1\nline 2\nline 3\n"
	go func() {
		defer w.Close()
		w.Write([]byte(testData))
	}()

	// Call the function
	result, err := cronlogger.ReadStdin()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if result != testData {
		t.Errorf("expected %q, got %q", testData, result)
	}
}

func Test_ReadPipeStdin_ReadError(t *testing.T) {
	// Save the original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe and close the read end to simulate a read error
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	// Close the read end to cause an error
	r.Close()
	os.Stdin = r

	// Write some data (this should fail)
	go func() {
		defer w.Close()
		w.Write([]byte("test"))
	}()

	// Call the function
	_, err = cronlogger.ReadStdin()
	if err == nil {
		t.Error("expected an error, got nil")
	}

	// Check that the error is not EOF
	if err == io.EOF {
		t.Error("expected a non-EOF error")
	}
}
