package main

import (
	"bufio"
	"flag"
	"fmt"
	"gotodo/internal"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// TODO: Add filtering
// TODO: Add coloring of deadline based on date
func main() {
	// Valid choices. First in each is default
	outputFormats := []string{"stdout", "color", "markdown"}
	sortColumns := []string{"deadline", "responsible", "tags"}

	// Command line variables
	var dir string
	flag.StringVar(&dir, "d", ".", "Directory to parse")

	var recurse bool
	flag.BoolVar(&recurse, "r", false, "Recurse subdirectories")

	var outputFormat string
	flag.StringVar(&outputFormat, "f", outputFormats[0], fmt.Sprintf("Output format %v", outputFormats))

	var useNerdfont bool
	flag.BoolVar(&useNerdfont, "u", true, "Use nerdfont symbols. May need to be false on older terminals")

	var sortBy string
	flag.StringVar(&sortBy, "s", sortColumns[0], fmt.Sprintf("Sort by [%v]", sortColumns))

	var sortDir bool
	flag.BoolVar(&sortDir, "a", true, "Sort ascending [true/false]")

	flag.Parse() // Parse input flags

	// Validate input date
	if !slices.Contains(outputFormats, outputFormat) {
		fmt.Fprintf(os.Stderr, "Incorrect format: %s. Allowed: %v\n", outputFormat, outputFormats)
		return
	}

	if !slices.Contains(sortColumns, sortBy) {
		fmt.Fprintf(os.Stderr, "Incorrect sort column: %s. Allowed: %v\n", sortBy, sortColumns)
		return
	}

	// Check terminal width
	w, _, err := term.GetSize(0)
	const wTotMin = 120
	const wDate = 12
	const wResp = 12
	const wTagsMin = 18
	const wMsgMin = wTotMin - wDate - wResp - wTagsMin // Use all available space

	if w < wTotMin {
		fmt.Fprintf(os.Stderr, "Terminal too narrow for output (%d < %d)\n", w, wTotMin)
		return
	}

	// Get todos
	todos, err := getTodos(dir, recurse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing todos: %v\n", err)
		return
	}

	// Print a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"T", "Activity", "Tags", "Responsible", "Deadline"})
	for _, todo := range todos {
		t.AppendRow(table.Row{todo.Status, todo.Msg, todo.Tags, todo.Responsible, todo.Deadline})
	}

	// Print correct format
	const ratioMsg = 0.7
	wTags := max(int(float32(1.0-ratioMsg)*float32(w-wDate)), wTagsMin)
	wMsg := w - wTags - wDate // Use as much space as possible for activity message

	// Helper to print todo
	todoTransformer := func(val any) string {
		if t, ok := val.(internal.TodoStatus); ok {
			if useNerdfont {
				switch t {
				case internal.StatusNotTodo:
					return ""
				case internal.StatusOpen:
					return " "
				case internal.StatusDone:
					return "󰄬"
				case internal.StatusCanceled:
					return "󰜺"
				}
			} else {
				switch t {
				case internal.StatusNotTodo:
					return "E"
				case internal.StatusOpen:
					return " "
				case internal.StatusDone:
					return "x"
				case internal.StatusCanceled:
					return "-"
				}
			}
		}

		return fmt.Sprintf("%v", val) // Fallback
	}

	// Helper to print string slices
	sliceTransformer := func(val any) string {
		if s, ok := val.([]string); ok {
			return strings.Join(s, ",")
		}

		return fmt.Sprintf("%v", val) // Fallback
	}

	// Helper to interpret dates
	dateTransformer := func(val any) string {
		if d, ok := val.(*time.Time); ok {
			if d == nil {
				return ""
			}

			if internal.IsTBD(*d) {
				return "TBD"
			}

			return d.Format("2006-01-02")
		}

		return fmt.Sprintf("%v", val)
	}

	// Configure table style
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:        "T",
			WidthMax:    3,
			Transformer: todoTransformer,
		},
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
			Name:        "Responsible",
			WidthMax:    wResp,
			Transformer: sliceTransformer,
		},
		{
			Name:        "Deadline",
			WidthMax:    wDate,
			Transformer: dateTransformer,
		},
	})

	// Sort the table
	var sortMode table.SortMode
	if sortDir {
		sortMode = table.Asc
	} else {
		sortMode = table.Dsc
	}

	var sortName string
	switch sortBy {
	case "deadline":
		sortName = "Deadline"
	case "tags":
		sortName = "Tags"
	case "responsiblep":
		sortName = "Resonsible"
	default:
		log.Panicf("invalid sort key")
	}

	t.SortBy([]table.SortBy{
		{Name: sortName, Mode: sortMode},
	})

	// Render to correct output
	switch outputFormat {
	case "stdout":
		t.Render()
	case "markdown":
		t.RenderMarkdown()
	case "color":
		{
			t.SetStyle(table.StyleColoredDark)
			t.Render()
		}
	}
}

// getTodos finds the files to check and calls parsing on them
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
