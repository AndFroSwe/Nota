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
		status:   NOT_TODO,
		deadline: time.Time{},
	}

	status, err := getStatus(s)
	if status == NOT_TODO {
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
