package internal

import (
	"slices"
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
		{"Empty string", "", "", false},
		{"Some input", "Some Input", "Some Input", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTodoLine(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got.raw != tt.want {
				t.Errorf("incorrect result parsing %s. Expected %v, got %v", tt.input, tt.want, got.raw)
			}
		})
	}
}

func TestMsg(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
		wantMsg string
		wantErr bool
	}{
		{"Empty String", "", "", "", false},
		{"Not a todo", "Not a todo", "Not a todo", "Not a todo", false},
		{"Todo with date", "- [ ] <260706> Todo w/ date", "- [ ] <260706> Todo w/ date", "Todo w/ date", false},
		{"Todo with date in UTF8", "- [ ] Ögli <260706> Tödå w/ date", "- [ ] Ögli <260706> Tödå w/ date", "Ögli Tödå w/ date", false},
		{"Todo with date in UTF8 and no space", "- [ ] Ögli<260706> Tödå w/ date", "- [ ] Ögli<260706> Tödå w/ date", "Ögli Tödå w/ date", false},
		{"Todo with tags", "- [x] A {tag1} tag", "- [x] A {tag1} tag", "A tag", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTodoLine(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got.raw != tt.wantRaw {
				t.Errorf("incorrect result parsing raw '%s'. Expected '%v', got '%v'", tt.input, tt.wantRaw, got.raw)
			}

			if got.msg != tt.wantMsg {
				t.Errorf("incorrect result parsing msg '%s'. Expected '%v', got '%v'", tt.input, tt.wantMsg, got.msg)
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
		{"status DONE", "- [x] Done todo", StatusDone, false},
		{"status OPEN", "- [ ] Open todo", StatusOpen, false},
		{"status CANCELED", "- [-] Canceled todo", StatusCanceled, false},
		{"Not a todo", "Just some text", StatusNotTodo, false},
		{"Late todo mark", "Text first - [x]", StatusNotTodo, false},
		{"Empty string", "", StatusNotTodo, false},
		{"UTF8 String", "- [x] Using ÅÄÖ", StatusDone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := getStatus(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("incorrect result parsing '%s'. Expected %v, got %v", tt.input, tt.want, got)
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTodoLine(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if got.deadline != tt.want {
				t.Errorf("incorrect result parsing '%s'. Expected %v, got %v", tt.input, tt.want, got.deadline)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"No tags", "- [x] No tags here", []string{}, false},
		{"One tag", "- [ ] One tag {tag1} here", []string{"tag1"}, false},
		{"Two tags", "- [ ] Two tags {tag1, tag2} here", []string{"tag1", "tag2"}, false},
		{"Tag with space", "- [ ] Spaced tag {tag space} in this", []string{"tag space"}, false},
		{"Two tag sections", "- [x] First {tag1} and second {tag2, tag 3}", []string{"tag1", "tag2", "tag 3"}, false},
		{"Not todo but has tags", "Not a todo but {my_tag} have tag", []string{"my_tag"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := getTags(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
				return
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("incorrect result parsing '%s': want '%v', got '%v'", tt.input, tt.want, got)
				return
			}
		})
	}
}

func TestParseResponsible(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"Not todo", "Not todo [[person]]", nil, false},
		{"No responsible", "- [ ] No responsible", nil, false},
		{"One responsible", "- [x] One Responsible [[af]]", []string{"af"}, false},
		{"Two responsible", "- [-] Two peeps [p1, fn ln]", []string{"p1", "fn ln"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := getResponsible(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("incorrect result parsing '%s': want '%v', got '%v'", tt.input, got, tt.want)
			}
		})
	}
}
