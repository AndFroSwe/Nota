package internal

import (
	"strings"
	"time"
)

type TodoStatus int

const (
	StatusNotTodo  TodoStatus = iota // Not a todo
	StatusOpen                       // Open todo
	StatusDone                       // Done!
	StatusCanceled                   // Not intended to be completed
	NumStatuses                      // Number of statuses, must keep this last
)

func GetAvailableStatuses() []string {
	return []string{"not-todo", "open", "done", "canceled"}
}

// A todo
type Todo struct {
	Raw      string     // Raw line input
	Msg      string     // Message line w/o metadata
	Status   TodoStatus // Status of the todo
	Deadline time.Time  // When the todo should be completed
	Tags     []string   // Togs in the todo
	Responsible []string // Responsible for executing todo
}

// extractedSurround is the return type when extracting tags, dates, responsibles or other types that
// are kept surrounded by symbols
type extractedSurround struct {
	contents string // Contents between open and close
	trimmed  string // String with open, close and contents removed
}

func getStatus(s string) (TodoStatus, string, error) {
	if strings.TrimSpace(s) == "" {
		return TodoStatus(StatusNotTodo), s, nil
	}

	// Define todo types
	const openTodo = "- [ ]"
	const doneTodo = "- [x]"
	const canceledTodo = "- [-]"

	// Find todo
	ss := strings.TrimSpace(s)
	if strings.HasPrefix(ss, openTodo) {
		return StatusOpen, ss[len(openTodo):], nil
	} else if strings.HasPrefix(ss, doneTodo) {
		return StatusDone, ss[len(doneTodo):], nil
	} else if strings.HasPrefix(ss, canceledTodo) {
		return StatusCanceled, ss[len(canceledTodo):], nil
	} else {
		return StatusNotTodo, s, nil
	}
}

func extractSurroundAndTrim(s string, openMarker string, closeMarker string) extractedSurround {
	// Find open marker
	start := strings.Index(s, openMarker)
	if start == -1 {
		return extractedSurround{"", s}
	}

	// Find close marker
	openSize := len(openMarker)
	closeSize := len(closeMarker)
	end := strings.Index(s[start+openSize:], closeMarker)
	if end == -1 {
		return extractedSurround{"", s}
	}

	// Found markers, extract contents
	extracted := s[start+openSize : start+end+openSize]

	return extractedSurround{
		contents: extracted,
		trimmed:  strings.TrimSpace(s[:start]) + " " + strings.TrimSpace(s[start+end+openSize+closeSize:]),
	}
}

func getDate(s string) (time.Time, string, error) {
	if strings.TrimSpace(s) == "" {
		return time.Time{}, s, nil
	}

	extracted := extractSurroundAndTrim(s, "<", ">")
	if extracted.contents == "" {
		return time.Time{}, s, nil
	}

	// Find date markers
	d, err := time.Parse("060102", extracted.contents)
	if err != nil {
		return time.Time{}, s, err
	}

	return d, extracted.trimmed, nil
}

// getTags extracts tags tag lists enclosed in curly braces separated with comma from a todo line
// Returns a string slice with tags or nil when no was found, a string with the tag syntax removed and an error
func getTags(s string) ([]string, string, error) {
	// Exit early on empty string
	if strings.TrimSpace(s) == "" {
		return nil, s, nil
	}

	// Extract tags
	var tags []string
	for {
		extracted := extractSurroundAndTrim(s, "{", "}")
		if extracted.contents == "" {
			break
		}

		tags = append(tags, strings.Split(extracted.contents, ",")...)
		s = extracted.trimmed
	}

	// Trim whitespace from tags
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}

	return tags, s, nil
}

// getResponsible takes a todo string and extracts persons responsible for executing a todo.
// Responsible are marked by [[ ]] and delimited by ,.
// Returns string slice with responsible, string with markers stripped out, error on error.
func getResponsible(s string) ([]string, string, error) {
	var responsibles []string
	for {
		extracted := extractSurroundAndTrim(s, "[[", "]]")
		if extracted.contents == "" {
			break
		}

		responsibles = append(responsibles, strings.Split(extracted.contents, ",")...)
		s = extracted.trimmed
	}

	// Trim whitespace
	for i := range responsibles {
		responsibles[i] = strings.TrimSpace(responsibles[i])
	}

	return responsibles, s, nil
}

func ParseTodoLine(s string) (Todo, error) {
	t := Todo{
		Raw:    s,
		Msg:    s,
		Status: StatusNotTodo,
	}

	// Get status and check if it is a todo
	status, s, err := getStatus(s)
	if err != nil {
		return t, err
	}

	// Don't parse if not todo
	if status == StatusNotTodo {
		return t, nil
	}

	// Was a todo, save status
	t.Status = status

	// Extract the date
	d, s, err := getDate(s)
	if err != nil {
		t.Msg = strings.TrimSpace(s) // Use current message
		return t, nil
	}

	// Had a date, save it
	t.Deadline = d

	// Extract the tags
	tags, s, err := getTags(s)
	if err != nil {
		t.Msg = strings.TrimSpace(s) // Use current message
	}

	// Save the values
	t.Tags = tags
	t.Msg = strings.TrimSpace(s)

	return t, nil
}
