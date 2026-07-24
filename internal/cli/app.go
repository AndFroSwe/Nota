package cli

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/andfroswe/nota/internal/todoitem"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// Table parameters
const (
	wTotMin   = 120                                                  // Min total width of terminal to present meaningful data
	wDate     = 12                                                   // Width of date column
	wResp     = 12                                                   // Width of responsible column
	wStatus   = 1                                                    // Width of status column
	wTagsMin  = 18                                                   // Min width for tags column
	wCols     = 5                                                    // Number of columns
	wOverhead = (wCols + 1) + (wCols-1)*2                            // Rendering overhead. Separators + padding
	wMsgMin   = wTotMin - wDate - wResp - wStatus - wTagsMin - wCols // Calculate the min width to use for the message column
)

// programOpts are the CLI program options
type programOpts struct {
	rootDir      string // Directory to start search for todo files in
	recurse      bool   // If true, recurse down from the root directory
	outputFormat string // How to present the table data
	useNerdfont  bool   // True if nerd fonts should be used
	sortBy       string // Column to sort by
	sortAsc      bool   // Sort ascending if true, descending otherwise
}

// Run is the main routine for using the CLI
//
// Returns error code, should be run like os.Exit(cli.Run())
func Run() int {
	opts, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing flags: ", err)
		return 1
	}

	// Check terminal width
	// Do this before parsing files to save time if terminal is too small anyway
	tableSize, err := calcTableSize()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error calculating tableSize: ", err)
		return 1
	}

	// Get todos
	todos, err := getTodos(opts.rootDir, opts.recurse)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing todos: ", err)
		return 1
	}

	// Create table with correct options
	t := createTable(opts, tableSize)
	for _, todo := range todos {
		t.AppendRow(table.Row{todo.Status, todo.Msg, todo.Tags, todo.Responsible, todo.Deadline})
	}

	// Render to correct output
	switch opts.outputFormat {
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

	return 0
}

// createTable creates a table with opts and tableSize and returns a table.Write with correct settings
func createTable(opts programOpts, tableSize *tableSize) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"T", "Activity", "Tags", "Responsible", "Deadline"})

	// Helper to print string slices
	sliceTransformer := func(val any) string {
		if s, ok := val.([]string); ok {
			return strings.Join(s, ",")
		}

		return fmt.Sprintf("%v", val) // Fallback
	}

	// Configure table style
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:        "T",
			WidthMax:    tableSize.wMaxStatus,
			Transformer: getTodoTransformer(opts),
		},
		{
			Name:     "Activity",
			WidthMax: tableSize.wMaxMsg,
		},
		{
			Name:        "Tags",
			WidthMax:    tableSize.wMaxTags,
			Transformer: sliceTransformer,
		},
		{
			Name:        "Responsible",
			WidthMax:    tableSize.wMaxResponsible,
			Transformer: sliceTransformer,
		},
		{
			Name:        "Deadline",
			WidthMax:    tableSize.wMaxDate,
			Transformer: getDateTransformer(),
		},
	})

	// Sort the table
	var sortMode table.SortMode
	if opts.sortAsc {
		sortMode = table.Asc
	} else {
		sortMode = table.Dsc
	}

	var sortName string
	switch opts.sortBy {
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
	return t
}

// getDateTransformer returns a transformer function to display dates in table
func getDateTransformer() func(val any) string {
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

	return dateTransformer
}

// getTodoTransformer takes programOpts and returns a transformer for displaying todo statuses
func getTodoTransformer(opts programOpts) func(val any) string {
	todoTransformer := func(val any) string {
		if t, ok := val.(internal.TodoStatus); ok {
			if opts.useNerdfont {
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
	return todoTransformer
}

// tableSize is the size informations for displaying the table
// Column width in glyphs
type tableSize struct {
	wMaxStatus      int
	wMaxMsg         int
	wMaxResponsible int
	wMaxTags        int
	wMaxDate        int
}

// calcTableSize calculates the size of the columns in the table based on parameters and the size of the terminal
func calcTableSize() (*tableSize, error) {
	terminalWidth, _, err := term.GetSize(0) // Get the current terminal size
	if err != nil {
		return nil, fmt.Errorf("error calculating tableSize: %v", err)
	}

	// Need enough space to give meaningful output
	if terminalWidth < wTotMin {
		return nil, fmt.Errorf("terminal too narrow for output (%d < %d)", terminalWidth, wTotMin)
	}

	tz := tableSize{
		wMaxDate:        wDate,
		wMaxResponsible: wResp,
		wMaxStatus:      wStatus,
	}
	const ratioMsg = 0.8                                                                    // The relative size of the dynamic columns to give to msg
	dynCols := terminalWidth - tz.wMaxDate - tz.wMaxResponsible - tz.wMaxStatus - wOverhead // Available columns for dynamics sizing
	tz.wMaxMsg = max(
		int(ratioMsg*float64(dynCols)), // Dynamic size
		wMsgMin,
	)
	tz.wMaxTags = max(
		dynCols-tz.wMaxMsg, // Left from msg size
		wTagsMin,           // Min allowed size
	)

	return &tz, nil
}

// parseFlags parses the input flags and returns programOpts if correct, error otherwise
func parseFlags() (programOpts, error) {
	opts := programOpts{}

	// Valid choices. First in each is default
	outputFormats := []string{"stdout", "color", "markdown"}
	sortColumns := []string{"deadline", "responsible", "tags"}

	// Command line variables
	flag.StringVar(&opts.rootDir, "d", ".", "Directory to parse")
	flag.BoolVar(&opts.recurse, "r", false, "Recurse subdirectories")
	flag.StringVar(&opts.outputFormat, "f", outputFormats[0], fmt.Sprintf("Output format %v", outputFormats))
	flag.BoolVar(&opts.useNerdfont, "u", true, "Use nerdfont symbols. May need to be false on older terminals")
	flag.StringVar(&opts.sortBy, "s", sortColumns[0], fmt.Sprintf("Sort by [%v]", sortColumns))
	flag.BoolVar(&opts.sortAsc, "a", true, "Sort ascending [true/false]")
	flag.Parse() // Parse input flags

	// Validate input date
	if !slices.Contains(outputFormats, opts.outputFormat) {
		fmt.Fprintf(os.Stderr, "Incorrect format: %s. Allowed: %v\n", opts.outputFormat, outputFormats)
		return programOpts{}, fmt.Errorf("incorrect output format. Want %s, got %v", outputFormats, opts.outputFormat)
	}

	if !slices.Contains(sortColumns, opts.sortBy) {
		fmt.Fprintf(os.Stderr, "Incorrect sort column: %s. Allowed: %v\n", opts.sortBy, sortColumns)
		return programOpts{}, fmt.Errorf("incorrect sort columns. Want %s, got %v", sortColumns, opts.sortBy)
	}

	return opts, nil
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

		todos, err = parseFile(path, todos) // Mutate parent todos

		return err
	})

	if err != nil {
		return nil, err
	}

	return todos, nil
}

// parseFile takes a path and pointer to todo slice and adds todos from file at path to the slice or returns an error
func parseFile(path string, todos []internal.Todo) ([]internal.Todo, error) {
	file, err := os.Open(path)
	if err != nil {
		return todos, fmt.Errorf("opening %s: %w", path, err)
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

		todos = append(todos, todo)
	}

	if err := scanner.Err(); err != nil {
		return todos, err
	}

	return todos, nil
}
