package diff

import (
	"reflect"
	"testing"
)

func TestLineByLine(t *testing.T) {
	tests := []struct {
		name  string
		text1 string
		text2 string
		want  []LineDiffResult
	}{
		{
			name:  "Identical",
			text1: "line1\nline2",
			text2: "line1\nline2",
			want: []LineDiffResult{
				{Type: "common", Line: "line1", Index: 1},
				{Type: "common", Line: "line2", Index: 2},
			},
		},
		{
			name:  "Addition",
			text1: "line1",
			text2: "line1\nline2",
			want: []LineDiffResult{
				{Type: "common", Line: "line1", Index: 1},
				{Type: "added", Line: "line2", Index: 2},
			},
		},
		{
			name:  "Deletion",
			text1: "line1\nline2",
			text2: "line1",
			want: []LineDiffResult{
				{Type: "common", Line: "line1", Index: 1},
				{Type: "removed", Line: "line2", Index: 2},
			},
		},
		{
			name:  "Modification",
			text1: "line1",
			text2: "line2",
			want: []LineDiffResult{
				{Type: "removed", Line: "line1", Index: 1},
				{Type: "added", Line: "line2", Index: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LineByLine(tt.text1, tt.text2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LineByLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
