package main

import (
	"bufio"
	"flag"
	"fmt"
	"gotodo/internal"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Command line variables
	var dir string
	flag.StringVar(&dir, "d", ".", "Directory to parse")

	var recurse bool
	flag.BoolVar(&recurse, "r", false, "Recurse subdirectories")

	flag.Parse()

	// Get todos
	todos, err := getTodos(dir, recurse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing todos: %v", err)
		return
	}

	fmt.Printf("Found %d todos\n", len(todos))
}

func getTodos(dir string, recurse bool) ([]internal.Todo, error) {
	recurse = false // Placeholder use
	var todos []internal.Todo
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		return parseFile(path, &todos)
	})

	if err != nil {
		return nil, err
	}

	return todos, nil
}

func parseFile(path string, todos *[]internal.Todo) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		todo, err := internal.ParseTodoLine(line)
		if err != nil {
			continue
		}

		if todo.Status == internal.StatusNotTodo {
			continue
		}

		*todos = append(*todos, todo)
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	return nil
}
