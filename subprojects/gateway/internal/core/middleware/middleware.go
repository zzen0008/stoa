package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

// OnCompletionFunc is a function that will be executed when a stream completes.
// It receives the full, buffered body of the response.
type OnCompletionFunc func(body []byte)

// ResponseMiddleware is a function that can modify the response and optionally
// return a function to be executed upon stream completion.
type ResponseMiddleware func(resp *http.Response) (OnCompletionFunc, error)

// StreamInterceptor is an io.ReadCloser that wraps the original response body,
// buffers the stream, and calls a finalizer function on EOF.
type StreamInterceptor struct {
	originalBody io.ReadCloser
	buffer       bytes.Buffer
	onCompletion OnCompletionFunc
}

// NewStreamInterceptor creates a new stream interceptor.
func NewStreamInterceptor(body io.ReadCloser, onCompletion OnCompletionFunc) *StreamInterceptor {
	return &StreamInterceptor{
		originalBody: body,
		onCompletion: onCompletion,
	}
}

// Read reads from the original body, writes to the buffer, and detects stream completion.
func (si *StreamInterceptor) Read(p []byte) (n int, err error) {
	n, err = si.originalBody.Read(p)
	if n > 0 {
		si.buffer.Write(p[:n])
	}
	if err == io.EOF {
		// Stream is complete, execute the finalizer
		if si.onCompletion != nil {
			// Run in a goroutine to avoid blocking the response flow.
			go si.onCompletion(si.buffer.Bytes())
		}
	}
	return n, err
}

// Close closes the original response body.
func (si *StreamInterceptor) Close() error {
	return si.originalBody.Close()
}

// Example: A simple logging middleware for responses.
func ElasticCompletionLogger(resp *http.Response) (OnCompletionFunc, error) {
	// This function is the middleware itself. It runs before the stream starts.
	// We can inspect headers here, for example.
	log.Printf("Response from downstream: status %d", resp.StatusCode)

	// We return a function that will be called ONLY when the stream is complete.
	onCompletion := func(body []byte) {
		// This is where you would send the data to Elasticsearch.
		log.Printf("STREAM COMPLETE. Logging to Elastic: %s", string(body))
	}

	return onCompletion, nil
}
