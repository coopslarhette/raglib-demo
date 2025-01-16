package api

import (
	"context"
	"log/slog"
	"raglib-demo/api/sse"
	"strconv"
	"strings"
)

type ChunkProcessor struct {
	citationBuffer strings.Builder
	codeBuffer     strings.Builder
	textBuffer     strings.Builder
	isCitation     bool
	isCodeBlock    bool
}

const (
	citationPrefixMarker  = "<cited>"
	citationPostfixMarker = "</cited>"
	codeBlockMarker       = "```"
)

// ProcessChunks should maybe be a standalone function instead of being a method of a struct
func (cp *ChunkProcessor) ProcessChunks(ctx context.Context, responseChan <-chan string, processedEventChan chan<- sse.Event) {
	defer func() {
		cp.flushRemainingBuffers(processedEventChan)
		close(processedEventChan)
	}()

	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				return
			}
			cp.processChunk(chunk, processedEventChan)
		case <-ctx.Done():
			return
		}
	}
}

func (cp *ChunkProcessor) flushRemainingBuffers(processedEventChan chan<- sse.Event) {
	if cp.codeBuffer.Len() > 0 {
		processedEventChan <- sse.NewCodeBlockEvent(cp.codeBuffer.String())
	} else if cp.citationBuffer.Len() > 0 {
		processedEventChan <- sse.NewTextEvent(cp.citationBuffer.String())
	}
}

func (cp *ChunkProcessor) processChunk(chunk string, processedEventChan chan<- sse.Event) {
	for _, char := range chunk {
		if cp.isCodeBlock {
			cp.processCodeBlockChar(char, processedEventChan)
		} else if cp.isCitation {
			cp.processCitationChar(char, processedEventChan)
		} else {
			cp.processTextChar(char, processedEventChan)
		}
	}
	cp.maybeFlushTextBufferTo(processedEventChan)
}

func (cp *ChunkProcessor) processCodeBlockChar(char rune, processedEventChan chan<- sse.Event) {
	cp.codeBuffer.WriteRune(char)
	if cp.codeBuffer.Len() < 4 {
		if char != '`' {
			cp.textBuffer.Write([]byte(cp.codeBuffer.String()))
			cp.codeBuffer.Reset()
			cp.isCodeBlock = false
		}
	} else if strings.HasSuffix(cp.codeBuffer.String(), codeBlockMarker) {
		processedEventChan <- sse.NewCodeBlockEvent(cp.codeBuffer.String())
		cp.codeBuffer.Reset()
		cp.isCodeBlock = false
	} else if !(strings.HasPrefix(codeBlockMarker, cp.codeBuffer.String()) || strings.HasPrefix(cp.codeBuffer.String(), codeBlockMarker)) {
		cp.textBuffer.Write([]byte(cp.codeBuffer.String()))
		cp.codeBuffer.Reset()
		cp.isCodeBlock = false
	}
}

func (cp *ChunkProcessor) processCitationChar(char rune, processedEventChan chan<- sse.Event) {
	cp.citationBuffer.WriteRune(char)
	if strings.HasSuffix(cp.citationBuffer.String(), citationPostfixMarker) {
		events := createCitationEvents(cp.citationBuffer)
		for _, event := range events {
			processedEventChan <- event
		}
		cp.citationBuffer.Reset()
		cp.isCitation = false
	} else if !(strings.HasPrefix(citationPrefixMarker, cp.citationBuffer.String()) || strings.HasPrefix(cp.citationBuffer.String(), citationPrefixMarker)) {
		cp.textBuffer.Write([]byte(cp.citationBuffer.String()))
		cp.citationBuffer.Reset()
		cp.isCitation = false
	}
}

// createCitationEvents supports multiple citation numbers within a single citation location ie "... <cited>2,3</cited> ..."
// but often it will just be returning a single citation event.
func createCitationEvents(citationBuffer strings.Builder) []sse.Event {
	citationStr := strings.TrimSuffix(citationBuffer.String(), citationPostfixMarker)
	citationStr = strings.TrimPrefix(citationStr, citationPrefixMarker)
	citationStr = strings.TrimSpace(citationStr)

	citations := strings.Split(citationStr, ",")
	events := make([]sse.Event, 0, len(citations))

	for _, citation := range citations {
		citation = strings.TrimSpace(citation)
		if citationNumber, err := strconv.Atoi(citation); err != nil {
			slog.Error("can't parse citation number from citation string", "citation string", citation, "err", err)
			events = append(events, sse.NewTextEvent(citation))
		} else {
			events = append(events, sse.NewCitationEvent(citationNumber))
		}
	}
	return events
}

func (cp *ChunkProcessor) processTextChar(char rune, processedEventChan chan<- sse.Event) {
	if char == '`' {
		cp.maybeFlushTextBufferTo(processedEventChan)
		cp.codeBuffer.WriteRune(char)
		cp.isCodeBlock = true
	} else if char == '<' {
		cp.maybeFlushTextBufferTo(processedEventChan)
		cp.citationBuffer.WriteRune(char)
		cp.isCitation = true
	} else {
		cp.textBuffer.WriteRune(char)
	}
}

func (cp *ChunkProcessor) maybeFlushTextBufferTo(processedEventChan chan<- sse.Event) {
	if cp.textBuffer.Len() > 0 {
		processedEventChan <- sse.NewTextEvent(cp.textBuffer.String())
		cp.textBuffer.Reset()
	}
}
