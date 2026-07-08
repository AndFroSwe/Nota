package internal

import (
	"errors"
	"strings"
	"time"
)

type TodoStatus int

const (
	StatusNotTodo TodoStatus = iota // Not a todo
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

func getStatus(s *string) (TodoStatus, error) {
	if strings.TrimSpace(*s) == "" {
		return TodoStatus(StatusNotTodo), errors.New("empty string")
	}

	ss := strings.TrimSpace(*s)
	if strings.HasPrefix(ss, "- [ ]") {
		return StatusOpen, nil
	} else if strings.HasPrefix(ss, "- [x]") {
		return StatusDone, nil
	} else if strings.HasPrefix(ss, "- [-]") {
		return StatusCanceled, nil
	} else {
		return StatusNotTodo, nil
	}
}

func getDate(s *string) (time.Time, error) {
	if strings.TrimSpace(*s) == "" {
		return time.Time{}, errors.New("empty string")
	}

	start := strings.Index(*s, "<")
	if start == -1 {
		return time.Time{}, errors.New("No <")
	}

	end := strings.Index((*s)[start+1:], ">")
	if end == -1 {
		return time.Time{}, errors.New("No >")
	}

	extracted := (*s)[start+1 : start+1+end]
	d, err := time.Parse("060102", extracted)
	if err != nil {
		return time.Time{}, err
	}

	return d, nil
}

func ParseTodoLine(s *string) (Todo, error) {
	t := Todo{
		raw:      *s,
		status:   StatusNotTodo,
		deadline: time.Time{},
	}

	status, err := getStatus(s)
	if status == StatusNotTodo {
		return t, nil
	}

	if err != nil {
		return t, err
	}
	t.status = status

	d, err := getDate(s)
	if err != nil {
		return t, nil
	}
	t.deadline = d

	return t, nil
}
