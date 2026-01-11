package tools

import (
	"bytes"
	"testing"
)

func TestGetLastNLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{
			name:  "get last 2 lines from 3",
			input: "line1\nline2\nline3\n",
			n:     2,
			want:  "line2\nline3\n",
		},
		{
			name:  "get more lines than available",
			input: "line1\nline2\n",
			n:     5,
			want:  "line1\nline2\n",
		},
		{
			name:  "empty input",
			input: " ",
			n:     1,
			want:  " ",
		},
		{
			name:  "single line no newline at end",
			input: "line1",
			n:     1,
			want:  "line1",
		},
		{
			name:  "last line is empty due to trailing newline",
			input: "line1\nline2\n",
			n:     1,
			want:  "line2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global state for the test
			messageMu.Lock()
			CurrentMessageLog = *bytes.NewBufferString(tt.input)
			messageMu.Unlock()

			got, err := GetLastNLines(tt.n)
			if err != nil {
				t.Fatalf("GetLastNLines() error = %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("GetLastNLines() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}
