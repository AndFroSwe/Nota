package internal

import "time"

type TodoStatus int

const (
	OPEN     TodoStatus = iota // Open todo
	DONE                       // Done!
	CANCELED                   // Not intended to be completed
	NOT_TODO                   // Not a todo
)

type Todo struct {
	raw      string
	status   TodoStatus
	deadline time.Time
}

func ParseTodoLine(s *string) (Todo, error) {
	return Todo{}, nil
}
