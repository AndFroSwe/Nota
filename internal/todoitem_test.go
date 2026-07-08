package internal

import (
	"testing"
	"time"
)

func TestRaw(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Empty string", "", "", true},
		{"Some input", "Some Input", "Some Input", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTodoLine(&tt.input)

			if err != nil && !tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got.raw != tt.want {
				t.Errorf("incorrect result parsing %s. Expected %v, got %v", tt.input, tt.want, got.raw)
			}
		})
	}
}

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
		{"Late todo mark", "Text first - [x]", NOT_TODO, false},
		{"Empty string", "", NOT_TODO, true},
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

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"Invalid month", "- [ ] <303102> Some text", time.Time{}, false},
		{"Invalid day", "- [ ] <300232> Some text", time.Time{}, false},
		{"Invalid date", "- [ ] <not_a_date> Some text", time.Time{}, false},
		{"Empty date", "- [ ] <> Some text", time.Time{}, false},
		{"Not a date", "- [ ] <not_a_date> Some text", time.Time{}, false},
		{"Not a todo", "<260102> Not a todo", time.Time{}, false},
		{"Open, no close", "Not a < date", time.Time{}, false},
		{"simple date", "- [ ] <260707> Some text", time.Date(2026, 07, 07, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			got, err := ParseTodoLine(&tt.input)

			if err != nil && !tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got.deadline != tt.want {
				t.Errorf("incorrect result parsing '%s'. Expected %v, got %v", tt.input, tt.want, got.deadline)
			}
		})
	}
}
