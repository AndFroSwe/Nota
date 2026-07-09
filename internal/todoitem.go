package internal

import (
	"errors"
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
}

func getStatus(s string) (TodoStatus, string, error) {
	if strings.TrimSpace(s) == "" {
		return TodoStatus(StatusNotTodo), s, errors.New("empty string")
	}

	// Define todo types
	const openTodo = "- [ ]"
	const doneTodo = "- [x]"
	const canceledTodo = "- [-]"

	ss := strings.TrimSpace(s)
	if strings.HasPrefix(ss, openTodo) {
		return StatusOpen, s[len(openTodo):], nil
	} else if strings.HasPrefix(ss, doneTodo) {
		return StatusDone, s[len(doneTodo):], nil
	} else if strings.HasPrefix(ss, canceledTodo) {
		return StatusCanceled, s[len(canceledTodo):], nil
	} else {
		return StatusNotTodo, s, nil
	}
}

func getDate(s string) (time.Time, string, error) {
	if strings.TrimSpace(s) == "" {
		return time.Time{}, s, errors.New("empty string")
	}

	const west = "<"
	const east = ">"
	// Find date markers
	start := strings.Index(s, west)
	if start == -1 {
		return time.Time{}, s, errors.New("No " + west)
	}

	end := strings.Index(s[start+1:], east)
	if end == -1 {
		return time.Time{}, s, errors.New("No " + east)
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

func ParseTodoLine(s string) (Todo, error) {
	t := Todo{
		raw:      s,
		msg:      s,
		status:   StatusNotTodo,
		deadline: time.Time{},
	}

	status, s, err := getStatus(s)
	if status == StatusNotTodo {
		return t, nil
	}

	if err != nil {
		return t, err
	}
	t.status = status

	d, s, err := getDate(s)
	if err != nil {
		t.msg = strings.TrimSpace(s) // Use current message
		return t, nil
	}
	t.deadline = d
	t.msg = strings.TrimSpace(s)

	return t, nil
}
