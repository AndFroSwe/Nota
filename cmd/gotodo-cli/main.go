package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
	"gotodo/internal"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	outputFormats := []string{"stdout", "color", "markdown"}

	// Command line variables
	var dir string
	flag.StringVar(&dir, "d", ".", "Directory to parse")

	var recurse bool
	flag.BoolVar(&recurse, "r", false, "Recurse subdirectories")

	var outputFormat string
	flag.StringVar(&outputFormat, "f", "stdout", fmt.Sprintf("Output format %v", outputFormats))

	flag.Parse()

	// Check input
	if !slices.Contains(outputFormats, outputFormat) {
		fmt.Fprintf(os.Stderr, "Incorrect format: %s. Allowed: %v\n", outputFormat, outputFormats)
		return
	}

	// Check terminal width
	w, _, err := term.GetSize(0)
	const wDate = 12
	const wTagsMin = 18
	const wMsgMin = 50

	wTotMin := wDate + wTagsMin + wMsgMin
	if w < wTotMin {
		fmt.Fprintf(os.Stderr, "Terminal too narrow for output (%d < %d)", w, wTotMin)
		return
	}

	// Get todos
	todos, err := getTodos(dir, recurse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing todos: %v", err)
		return
	}

	// Print a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Activity", "Tags", "Deadline"})
	for _, todo := range todos {
		t.AppendRow(table.Row{todo.Msg, todo.Tags, todo.Deadline.Format("2006-01-02")})
	}

	// Print correct format
	const ratioMsg = 0.7
	wTags := max(int(float32(1.0-ratioMsg)*float32(w-wDate)), wTagsMin)
	wMsg := w - wTags - wDate // Use as much space as possible

	sliceTransformer := func(val any) string {
		if s, ok := val.([]string); ok {
			return strings.Join(s, ",")
		}

		return fmt.Sprintf("%v", val) // Fallback
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:     "Activity",
			WidthMax: wMsg,
		},
		{
			Name:        "Tags",
			WidthMax:    wTags,
			Transformer: sliceTransformer,
		},
		{
			Name:     "Deadline",
			WidthMax: wDate,
		},
	})
	switch outputFormat {
	case "stdout":
		t.Render()
	case "markdown":
		t.RenderMarkdown()
	case "color":
		{
			t.SetStyle(table.StyleColoredBlackOnBlueWhite)
			t.Render()
		}
	}
}

func getTodos(dir string, recurse bool) ([]internal.Todo, error) {
	var todos []internal.Todo
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if !recurse && path != dir {
				return filepath.SkipDir
			}
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
