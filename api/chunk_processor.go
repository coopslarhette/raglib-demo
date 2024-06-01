package api

import (
	"fmt"
	"raglib/api/sse"
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
func (cp *ChunkProcessor) ProcessChunks(responseChan <-chan string, bufferedChunkChan chan<- sse.Event) {
	for chunk := range responseChan {
		for _, char := range chunk {
			if cp.isCodeBlock {
				cp.processCodeBlockChar(char, bufferedChunkChan)
			} else if cp.isCitation {
				cp.processCitationChar(char, bufferedChunkChan)
			} else {
				cp.processTextChar(char, bufferedChunkChan)
			}
		}
		cp.maybeFlushTextBuffer(bufferedChunkChan)
	}

	if cp.codeBuffer.Len() > 0 {
		bufferedChunkChan <- sse.NewCodeBlockEvent(cp.codeBuffer.String())
	} else if cp.citationBuffer.Len() > 0 {
		bufferedChunkChan <- sse.NewTextEvent(cp.citationBuffer.String())
	}

	close(bufferedChunkChan)
}

func (cp *ChunkProcessor) processCodeBlockChar(char rune, bufferedChunkChan chan<- sse.Event) {
	cp.codeBuffer.WriteRune(char)
	if strings.HasSuffix(cp.codeBuffer.String(), codeBlockMarker) {
		bufferedChunkChan <- sse.NewCodeBlockEvent(cp.codeBuffer.String())
		cp.codeBuffer.Reset()
		cp.isCodeBlock = false
	}
}

func (cp *ChunkProcessor) processCitationChar(char rune, bufferedChunkChan chan<- sse.Event) {
	cp.citationBuffer.WriteRune(char)
	if strings.HasSuffix(cp.citationBuffer.String(), citationPostfixMarker) {
		bufferedChunkChan <- createCitationEvent(cp.citationBuffer)
		cp.citationBuffer.Reset()
		cp.isCitation = false
	} else if !(strings.HasPrefix(citationPrefixMarker, cp.citationBuffer.String()) || strings.HasPrefix(cp.citationBuffer.String(), citationPrefixMarker)) {
		cp.textBuffer.Write([]byte(cp.citationBuffer.String()))
		cp.citationBuffer.Reset()
		cp.isCitation = false
	}
}

func createCitationEvent(citationBuffer strings.Builder) sse.Event {
	citationStr := strings.TrimSuffix(citationBuffer.String(), citationPostfixMarker)
	citationStr = strings.TrimPrefix(citationStr, citationPrefixMarker)
	citationStr = strings.TrimSpace(citationStr)
	if citationNumber, err := strconv.Atoi(citationStr); err != nil {
		fmt.Printf("Invalid citation number text in between citation marker XML tags: %s\n", citationStr)
		return sse.NewTextEvent(citationStr)
	} else {
		return sse.NewCitationEvent(citationNumber)
	}
}

func (cp *ChunkProcessor) processTextChar(char rune, bufferedChunkChan chan<- sse.Event) {
	codeBlockStart := strings.Index(cp.textBuffer.String()+string(char), codeBlockMarker)
	if codeBlockStart != -1 {
		text := cp.textBuffer.String() + string(char)
		cp.textBuffer.Reset()
		// Flush any text before code block that is currently buffered
		if codeBlockStart > 0 {
			bufferedChunkChan <- sse.NewTextEvent(text[:codeBlockStart])
		}
		cp.codeBuffer.WriteString(text[codeBlockStart:])
		cp.isCodeBlock = true
	} else if char == '<' {
		cp.maybeFlushTextBuffer(bufferedChunkChan)
		cp.citationBuffer.WriteRune(char)
		cp.isCitation = true
	} else {
		cp.textBuffer.WriteRune(char)
	}
}

func (cp *ChunkProcessor) maybeFlushTextBuffer(bufferedChunkChan chan<- sse.Event) {
	if cp.textBuffer.Len() > 0 {
		bufferedChunkChan <- sse.NewTextEvent(cp.textBuffer.String())
		cp.textBuffer.Reset()
	}
}
