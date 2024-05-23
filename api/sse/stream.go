package sse

import (
	"encoding/json"
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

	return nil
}

func (s *Stream) Write(e Event) error {
	marshalled, err := json.Marshal(e.Data)
	if err != nil {
		return fmt.Errorf("error marshalling event data: %v", err)
	}

	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", e.EventType, marshalled); err != nil {
		return fmt.Errorf("error writing to streaming: %v", err)
	}

	s.flusher.Flush()
	return nil
}

func (s *Stream) Error(clientErrorMessage string) {
	fmt.Fprintf(s.w, "event: error\ndata: %s\n\n", clientErrorMessage)
	s.flusher.Flush()
}

type Event struct {
	EventType string
	Data      interface{}
}

func NewTextEvent(text string) Event {
	return Event{EventType: "text", Data: text}
}

func NewCitationEvent(citationNumber int) Event {
	return Event{EventType: "citation", Data: citationNumber}
}

func NewErrorEvent(errorMessage string) Event {
	return Event{EventType: "error", Data: errorMessage}
}
