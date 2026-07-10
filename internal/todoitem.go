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
)

type Todo struct {
	raw      string     // Raw line input
	msg      string     // Message line w/o metadata
	status   TodoStatus // Status of the todo
	deadline time.Time  // When the todo should be completed
	tags     []string   // Togs in the todo
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

func getDate(s string) (time.Time, string, error) {
	if strings.TrimSpace(s) == "" {
		return time.Time{}, s, nil
	}

	const west = "<"
	const east = ">"
	// Find date markers
	start := strings.Index(s, west)
	if start == -1 {
		return time.Time{}, s, nil
	}

	end := strings.Index(s[start+1:], east)
	if end == -1 {
		return time.Time{}, s, nil
	}

	extracted := s[start+1 : start+1+end]
	d, err := time.Parse("060102", extracted)
	if err != nil {
		return time.Time{}, s, err
	}

	startMsg := strings.TrimSpace(s[:start])
	endMsg := strings.TrimSpace(s[start+end+2:])
	msg := startMsg + " " + endMsg // Remove extracted part

	return d, msg, nil
}

func getTags(s string) ([]string, string, error) {
	if strings.TrimSpace(s) == "" {
		return []string{}, s, nil
	}

	const west = "{"
	const east = "}"

	start := strings.Index(s, west) // Find start of tag marker
	if start == -1 {
		return []string{}, s, nil
	}

	end := strings.Index(s[start+1:], east) // Find end of tag marker
	if end == -1 {
		return []string{}, s, nil
	}

	// Found a marker, extract tags
	const sep = "," // Tag separator
	tags := strings.Split(s[start+1:start+end+1], sep)
	// Trim whitespace from tags
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}

	// Remove tag from msg
	s = strings.TrimSpace(s[:start]) + " " + strings.TrimSpace(s[start+end+2:])

	return tags, s, nil
}

func ParseTodoLine(s string) (Todo, error) {
	t := Todo{
		raw:    s,
		msg:    s,
		status: StatusNotTodo,
	}

	// Get status and check if it is a todo
	status, s, err := getStatus(s)
	if err != nil {
		return t, err
	}

	if status == StatusNotTodo {
		return t, nil
	}

	// Was a todo, save status
	t.status = status

	// Extract the date
	d, s, err := getDate(s)
	if err != nil {
		t.msg = strings.TrimSpace(s) // Use current message
		return t, nil
	}

	// Had a date, save it
	t.deadline = d

	// Extract the tags
	tags, s, err := getTags(s)
	if err != nil {
		t.msg = strings.TrimSpace(s) // Use current message
	}

	// Save the values
	t.tags = tags
	t.msg = strings.TrimSpace(s)

	return t, nil
}
