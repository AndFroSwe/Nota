package internal

import (
	"errors"
	"strings"
	"time"
)

type TodoStatus int

const (
	NOT_TODO TodoStatus = iota // Not a todo
	OPEN                       // Open todo
	DONE                       // Done!
	CANCELED                   // Not intended to be completed
)

type Todo struct {
	raw      string
	status   TodoStatus
	deadline time.Time
}

func getStatus(s *string) (TodoStatus, error) {
	if strings.TrimSpace(*s) == "" {
		return TodoStatus(NOT_TODO), errors.New("empty string")
	}

	ss := strings.TrimSpace(*s)
	if strings.HasPrefix(ss, "- [ ]") {
		return OPEN, nil
	} else if strings.HasPrefix(ss, "- [x]") {
		return DONE, nil
	} else if strings.HasPrefix(ss, "- [-]") {
		return CANCELED, nil
	} else {
		return NOT_TODO, nil
	}
}

func ParseTodoLine(s *string) (Todo, error) {
	t := Todo{
		raw:      *s,
		status:   NOT_TODO,
		deadline: time.Now(),
	}

	status, err := getStatus(s)
	if err != nil {
		return t, err
	}
	t.status = status

	return t, nil
}
