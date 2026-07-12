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

			if got.Raw != tt.want {
				t.Errorf("incorrect result parsing %s. Expected %v, got %v", tt.input, tt.want, got.Raw)
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

			if got.Raw != tt.wantRaw {
				t.Errorf("incorrect result parsing raw '%s'. Expected '%v', got '%v'", tt.input, tt.wantRaw, got.Raw)
			}

			if got.Msg != tt.wantMsg {
				t.Errorf("incorrect result parsing msg '%s'. Expected '%v', got '%v'", tt.input, tt.wantMsg, got.Msg)
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
		{"Not a todo w/ tags", "Not todo [[person]]", StatusNotTodo, false},
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
		name     string
		input    string
		wantNil  bool      // If  a nil pointer is expected
		wantDate time.Time // If we expect a certain time
		wantErr  bool
	}{
		{"Invalid month", "- [ ] <303102> Some text", true, time.Time{}, true},
		{"Invalid day", "- [ ] <300232> Some text", true, time.Time{}, true},
		{"Invalid date", "- [ ] <not_a_date> Some text", true, time.Time{}, true},
		{"Empty date", "- [ ] <> Some text", true, time.Time{}, true},
		{"Not a date", "- [ ] <not_a_date> Some text", true, time.Time{}, true},
		{"Not a todo", "<260102> Not a todo", true, time.Time{}, false},
		{"Open, no close", "Not a < date", true, time.Time{}, true},
		{"simple date", "- [ ] <260707> Some text", false, time.Date(2026, 07, 07, 0, 0, 0, 0, time.UTC), false},
		{"TBD date", "- [ ] <?> Date to be decided", false, time.Time{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := getDate(tt.input)

			// Check for error
			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			// Check for nil
			if (got == nil) != tt.wantNil {
				t.Errorf("incorrect nil when parsing '%s': want '%v', got '%v'", tt.input, tt.wantNil, (got == nil))
			}

			// Check for value
			if *got != tt.wantDate {
				t.Errorf("incorrect result parsing '%s'. Expected %v, got %v", tt.input, tt.wantDate, got)
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
		want    []string // Responsibles
		wantMsg string   // Trimmed message string
		wantErr bool
	}{
		{"No responsible", "- [ ] No responsible", nil, "- [ ] No responsible", false},
		{"One responsible", "- [x] One Responsible [[af]]", []string{"af"}, "- [x] One Responsible", false},
		{"Two responsible", "- [-] Two peeps [[p1, fn ln]] after", []string{"p1", "fn ln"}, "- [-] Two peeps after", false},
		{"With UTF8", "- [ ] A [[LÅE]] responsible with UTF8", []string{"LÅE"}, "- [ ] A responsible with UTF8", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotMsg, err := getResponsible(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error parsing '%s': %v", tt.input, err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("incorrect result parsing '%s': want '%v', got '%v'", tt.input, tt.want, got)
			}

			if gotMsg != tt.wantMsg {
				t.Errorf("incorrect msg parsing '%s': want '%v', got '%v'", tt.input, tt.wantMsg, gotMsg)
			}
		})
	}
}

// TestAvailableStatuses checks that available statuses match with print function
func TestAvailableStatuses(t *testing.T) {
	got := len(GetAvailableStatuses())
	want := int(NumStatuses)

	if got != want {
		t.Errorf("incorrect GetAvailableStatuses: want %d, got %d", want, got)
	}

}

// TODO: Add full parse test
