package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestStreamInterceptor ensures the body is read correctly and the onCompletion function is called.
func TestStreamInterceptor(t *testing.T) {
	originalBody := "Hello, World!"
	bodyReader := io.NopCloser(strings.NewReader(originalBody))

	var wg sync.WaitGroup
	wg.Add(1)

	var completedBody []byte
	onCompletion := func(body []byte) {
		completedBody = body
		wg.Done()
	}

	interceptor := NewStreamInterceptor(bodyReader, onCompletion)

	// Read the entire body through the interceptor
	readBody, err := io.ReadAll(interceptor)
	if err != nil {
		t.Fatalf("Failed to read from interceptor: %v", err)
	}

	if string(readBody) != originalBody {
		t.Errorf("Interceptor did not read the correct body, got: %q, want: %q", string(readBody), originalBody)
	}

	// Wait for the onCompletion function to be called
	wg.Wait()

	if string(completedBody) != originalBody {
		t.Errorf("onCompletion received incorrect body, got: %q, want: %q", string(completedBody), originalBody)
	}
}

// TestElasticCompletionLogger tests the logging output of the example middleware.
func TestElasticCompletionLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	// Create a mock response
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	// Call the middleware
	onCompletion, err := ElasticCompletionLogger(resp)
	if err != nil {
		t.Fatalf("Middleware returned an unexpected error: %v", err)
	}

	// Check the initial log message
	expectedInitialLog := `level=info msg="Response from downstream: status 200"`
	if !strings.Contains(buf.String(), expectedInitialLog) {
		t.Errorf("Initial log was incorrect, got: %q, want to contain: %q", buf.String(), expectedInitialLog)
	}

	// Reset buffer and call the completion function
	buf.Reset()
	completionBody := []byte("This is the completion body.")
	onCompletion(completionBody)

	// Check the completion log message
	expectedCompletionLog := `level=info msg="STREAM COMPLETE. Logging to Elastic: This is the completion body."`
	if !strings.Contains(buf.String(), expectedCompletionLog) {
		t.Errorf("Completion log was incorrect, got: %q, want to contain: %q", buf.String(), expectedCompletionLog)
	}
}
