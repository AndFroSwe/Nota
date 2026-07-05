package internal

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TodoStatus
		wantErr bool
	}{
		{"status DONE", "- [x] Done todo", DONE, false},
		{"status OPEN", "- [ ] Open todo", OPEN, false},
		{"status CANCELED", "- [-] Canceled todo", CANCELED, false},
		{"Not a todo", "Just some text", NOT_TODO, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTodoLine(&tt.input)

			if err != nil && !tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got.status != tt.want {
				t.Errorf("incorrect result parsing '%s'. Expected %v, got %v", tt.input, tt.want, got.status)
			}
		})
	}

}
