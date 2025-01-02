package data

import (
	"regexp"
	"testing"
)

func TestGenerateNoteID(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Title with spaces",
			args: args{title: "My new note"},
			want: "expected-id-1",
		},
		{
			name: "Empty title",
			args: args{title: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateNoteID(tt.args.title)
			if tt.args.title == "" {
				if !regexp.MustCompile(`^\d{10}-[A-Z]{4}$`).MatchString(got) {
					t.Errorf("GenerateNoteID() = %v, does not match expected format for empty title", got)
				}
			} else {
				if !regexp.MustCompile(`^\d{10}-[a-z0-9-]+$`).MatchString(got) {
					t.Errorf("GenerateNoteID() = %v, does not match expected format", got)
				}
			}
		})
	}
}
