package api

import (
	"fmt"
	"net/http"
)

type Stream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewStream(w http.ResponseWriter) Stream {
	return Stream{w: w}
}

// Establish establishes the SSE connection via writing the appropriate headers and flushing the response writer
func (s *Stream) Establish() error {
	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")

	// Flush the response writer to establish the SSE connection
	f, ok := s.w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}
	f.Flush()
	s.flusher = f
	println("attaching flushed", s.flusher)
	return nil
}

func (s *Stream) Write(data interface{}) error {
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("error writing to streaming: %v", err)
	}
	if s.flusher == nil {
		println("flushing is nil")
	}
	s.flusher.Flush()
	return nil
}
