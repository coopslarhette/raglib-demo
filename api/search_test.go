package api

import (
	"raglib/api/sse"
	"testing"
)

func TestProcessAndBufferChunks(t *testing.T) {
	testCases := []struct {
		name           string
		inputChunks    []string
		expectedOutput []sse.Event
	}{
		{
			name:        "Citation in the middle",
			inputChunks: []string{"This is a sample text with a citation <cited>1</cited> in the middle."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("This is a sample text with a citation "),
				sse.NewCitationEvent(1),
				sse.NewTextEvent(" in the middle."),
			},
		},
		{
			name:        "Multiple citations",
			inputChunks: []string{"Another example with multiple citations <cited>2</cited> and <cited>3</cited>."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("Another example with multiple citations "),
				sse.NewCitationEvent(2),
				sse.NewTextEvent(" and "),
				sse.NewCitationEvent(3),
				sse.NewTextEvent("."),
			},
		},
		{
			name:        "Partial citation spanning multiple chunks",
			inputChunks: []string{"An edge case with a partial citation <ci", "ted>4</cited> that spans multiple chunks."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("An edge case with a partial citation "),
				sse.NewCitationEvent(4),
				sse.NewTextEvent(" that spans multiple chunks."),
			},
		},
		{
			name:        "Citation at the end of a chunk",
			inputChunks: []string{"A case with a citation at the end of a chunk <cited>5</cited>", " and some text after it."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("A case with a citation at the end of a chunk "),
				sse.NewCitationEvent(5),
				sse.NewTextEvent(" and some text after it."),
			},
		},
		{
			name:        "Incomplete citation at the end",
			inputChunks: []string{"A case with an incomplete citation at the end <cit"},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("A case with an incomplete citation at the end "),
				sse.NewTextEvent("<cit"),
			},
		},
		{
			name:        "Code block",
			inputChunks: []string{"A code block example:", "```python\n", "print('Hello, World!')\n", "```", "End of code block."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("A code block example:"),
				sse.NewCodeBlockEvent("```python\nprint('Hello, World!')\n```"),
				sse.NewTextEvent("End of code block."),
			},
		},
		{
			name:        "Code block spanning multiple chunks",
			inputChunks: []string{"Another code block ", "example:", "```java\n", "System.out.println(\"Hello, ", "World!\");\n", "```", "End of code block."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("Another code block "),
				sse.NewTextEvent("example:"),
				sse.NewCodeBlockEvent("```java\nSystem.out.println(\"Hello, World!\");\n```"),
				sse.NewTextEvent("End of code block."),
			},
		},
		{
			name:        "Code block with citation",
			inputChunks: []string{"A code block with a citation:", "```javascript\n", "console.log('Citation: <cited>6</cited>');\n", "```", "End of code block."},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("A code block with a citation:"),
				sse.NewCodeBlockEvent("```javascript\nconsole.log('Citation: <cited>6</cited>');\n```"),
				sse.NewTextEvent("End of code block."),
			},
		},
		{
			name:        "Incomplete code block at the end",
			inputChunks: []string{"An incomplete code block at the end:", "```"},
			expectedOutput: []sse.Event{
				sse.NewTextEvent("An incomplete code block at the end:"),
				sse.NewCodeBlockEvent("```"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			responseChan := make(chan string)
			bufferedChunkChan := make(chan sse.Event)

			go func() {
				for _, chunk := range tc.inputChunks {
					responseChan <- chunk
				}
				close(responseChan)
			}()

			p := ChunkProcessor{}
			go p.ProcessChunks(responseChan, bufferedChunkChan)

			var outputEvents []sse.Event
			for event := range bufferedChunkChan {
				outputEvents = append(outputEvents, event)
			}

			if len(outputEvents) != len(tc.expectedOutput) {
				t.Fatalf("Unexpected number of output events. Got: %d, Expected: %d", len(outputEvents), len(tc.expectedOutput))
			}

			for i, event := range outputEvents {
				if event.EventType != tc.expectedOutput[i].EventType {
					t.Errorf("Unexpected output event type. Got: %v, Expected: %v", event.EventType, tc.expectedOutput[i].EventType)
				}
				if event.Data != tc.expectedOutput[i].Data {
					t.Errorf("Unexpected output event data. Got: %+v, Expected: %+v", event.Data, tc.expectedOutput[i].Data)
				}
			}
		})
	}
}
