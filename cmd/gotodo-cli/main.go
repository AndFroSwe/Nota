package main

import (
	"bufio"
	"fmt"
	"gotodo/internal"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "."

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
		fmt.Printf("error walking directory: %v", err)
	}

	fmt.Printf("Found %d todos\n", len(todos))

	for i, t := range todos {
		fmt.Printf("%d: %s\n", i, t.Raw)
	}
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
