package api

import (
	"testing"
)

func TestProcessAndBufferChunks(t *testing.T) {
	testCases := []struct {
		name           string
		inputChunks    []string
		expectedOutput []string
	}{
		{
			name:           "Citation in the middle",
			inputChunks:    []string{"This is a sample text with a citation <cited>1</cited> in the middle."},
			expectedOutput: []string{"This is a sample text with a citation ", "<cited>1</cited>", " in the middle."},
		},
		{
			name:           "Multiple citations",
			inputChunks:    []string{"Another example with multiple citations <cited>2</cited> and <cited>3</cited>."},
			expectedOutput: []string{"Another example with multiple citations ", "<cited>2</cited>", " and ", "<cited>3</cited>", "."},
		},
		{
			name:           "Partial citation spanning multiple chunks",
			inputChunks:    []string{"An edge case with a partial citation <ci", "ted>4</cited> that spans multiple chunks."},
			expectedOutput: []string{"An edge case with a partial citation ", "<cited>4</cited>", " that spans multiple chunks."},
		},
		{
			name:           "Citation at the end of a chunk",
			inputChunks:    []string{"A case with a citation at the end of a chunk <cited>5</cited>", " and some text after it."},
			expectedOutput: []string{"A case with a citation at the end of a chunk ", "<cited>5</cited>", " and some text after it."},
		},
		{
			name:           "Incomplete citation at the end",
			inputChunks:    []string{"A case with an incomplete citation at the end <cit"},
			expectedOutput: []string{"A case with an incomplete citation at the end ", "<cit"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			responseChan := make(chan string)
			bufferedChunkChan := make(chan string)

			go func() {
				for _, chunk := range tc.inputChunks {
					responseChan <- chunk
				}
				close(responseChan)
			}()

			go processAndBufferChunks(responseChan, bufferedChunkChan)

			var outputChunks []string
			for chunk := range bufferedChunkChan {
				outputChunks = append(outputChunks, chunk)
			}

			if len(outputChunks) != len(tc.expectedOutput) {
				t.Fatalf("Unexpected number of output chunks. Got: %d, Expected: %d", len(outputChunks), len(tc.expectedOutput))
			}

			for i, chunk := range outputChunks {
				if chunk != tc.expectedOutput[i] {
					t.Errorf("Unexpected output chunk. Got: %q, Expected: %q", chunk, tc.expectedOutput[i])
				}
			}
		})
	}
}
