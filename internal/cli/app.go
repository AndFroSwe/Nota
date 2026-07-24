package cli

import (
	"bufio"
	"flag"
	"fmt"
	todoitem "github.com/andfroswe/nota/internal/todoitem"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

// Table parameters
const (
	minTotalWidth       = 120 // Min total width of terminal to present meaningful data
	maxDateWidth        = 12  // Width of date column
	maxResponisbleWidth = 12  // Width of responsible column
	maxStatusWidth      = 1   // Width of status column
	minTagsWidth        = 18  // Min width for tags column
	maxFileWidth        = 30  // Min width for todo file
)

// Allowed columns to sort by
var sortColumns = []string{"status", "deadline", "responsible", "tags"}

// Allowed output format
var outputFormats = []string{"stdout", "color", "markdown"}

// programOpts are the CLI program options
type programOpts struct {
	rootDir        string   // Directory to start search for todo files in
	recurse        bool     // If true, recurse down from the root directory
	outputFormat   string   // How to present the table data
	useNerdfont    bool     // True if nerd fonts should be used
	sortBy         string   // Column to sort by
	sortAsc        bool     // Sort ascending if true, descending otherwise
	filterByStatus []string // Only show tasks with these statuses
	listFile       bool     // If true, display column with the file task is in
	showVersion    bool     // If true, show version and quit
}

// tableSize is the size informations for displaying the table
// Column width in glyphs
type tableSize struct {
	wMaxStatus      int
	wMinMsg         int
	wMaxResponsible int
	wMinTags        int
	wMaxDate        int
	wMaxFile        int
}

// Run is the main routine for using the CLI
//
// Returns error code, should be run like os.Exit(cli.Run())
func Run() int {
	opts := parseFlags()

	// If version switch given, display version and exit
	if opts.showVersion {
		fmt.Printf("NoTa version %s\n", getVersion())
		return 0
	}

	// Check terminal width
	// Do this before parsing files to save time if terminal is too small anyway
	tableSize, err := calcTableSize(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error calculating tableSize: ", err)
		return 1
	}

	// Create table with correct options
	t, err := createTable(opts, tableSize)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error in createTable: ", err)
		return 1
	}

	// Get todos
	todos, err := getTodos(opts.rootDir, opts.recurse)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing todos: ", err)
		return 1
	}

	// Process data
	tableData, err := filterData(opts, todos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}

	// Add the data to the table
	for _, todo := range tableData {
		row := table.Row{todo.Status, todo.Msg, todo.Tags, todo.Responsible, todo.Deadline}
		if opts.listFile {
			row = append(row, todo.File)
		}

		t.AppendRow(row)
	}

	// Render to correct output
	if !slices.Contains(outputFormats, opts.outputFormat) {
		fmt.Fprintf(os.Stderr, "incorrect outputFormat: Want %v, got %s\n", outputFormats, opts.outputFormat)
		return 1
	}

	switch opts.outputFormat {
	case "stdout":
		t.Render()
	case "markdown":
		t.RenderMarkdown()
	case "color":
		t.SetStyle(table.StyleColoredDark)
		t.Render()
	}

	return 0
}

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return info.Main.Version
}

func filterData(opts programOpts, todos []todoitem.Todo) ([]todoitem.Todo, error) {
	filteredData := make([]todoitem.Todo, 0, len(todos))

	for _, status := range opts.filterByStatus {
		if !slices.Contains(todoitem.GetAvailableStatuses(), status) {
			return nil, fmt.Errorf("invalid filter status. Want %v, got %s\n", todoitem.GetAvailableStatuses(), status)
		}
	}

	for _, todo := range todos {
		if slices.Contains(opts.filterByStatus, todoitem.ToString(todo.Status)) {
			filteredData = append(filteredData, todo)
		}
	}
	return filteredData, nil
}

// createTable creates a table with opts and tableSize and returns a table.Write with correct settings
func createTable(opts programOpts, tableSize tableSize) (table.Writer, error) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Configure table style
	statusColumnName := "T" // Set this to a one letter name to match icon length
	headers := table.Row{statusColumnName, "Activity", "Tags", "Responsible", "Deadline"}
	columns := []table.ColumnConfig{
		{
			Name:        statusColumnName,
			WidthMax:    tableSize.wMaxStatus,
			Transformer: getTodoTransformer(opts),
		},
		{
			Name:     "Activity",
			WidthMax: tableSize.wMinMsg,
		},
		{
			Name:        "Tags",
			WidthMax:    tableSize.wMinTags,
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
			Transformer: dateTransformer,
		},
	}

	// Add file column if needed
	if opts.listFile {
		headers = append(headers, "File")
		columns = append(columns, table.ColumnConfig{
			Name:     "File",
			WidthMax: tableSize.wMaxFile,
		})

	}

	// Set table config
	t.AppendHeader(headers)
	t.SetColumnConfigs(columns)

	// Sort the table
	var sortMode table.SortMode
	if opts.sortAsc {
		sortMode = table.Asc
	} else {
		sortMode = table.Dsc
	}

	if !slices.Contains(sortColumns, opts.sortBy) {
		fmt.Fprintf(os.Stderr, "Incorrect sort column: %s. Allowed: %v\n", opts.sortBy, sortColumns)
		return nil, fmt.Errorf("incorrect sort columns. Want %s, got %v", sortColumns, opts.sortBy)
	}

	var sortName string
	switch opts.sortBy {
	case "status":
		sortName = statusColumnName
	case "deadline":
		sortName = "Deadline"
	case "tags":
		sortName = "Tags"
	case "responsible":
		sortName = "Responsible"
	default:
		return nil, fmt.Errorf("incorrect sort key. Want %v, got %s", sortColumns, opts.sortBy)
	}

	t.SortBy([]table.SortBy{
		{Name: sortName, Mode: sortMode},
	})

	return t, nil
}

// dateTransformer transforms a date to desired format
var dateTransformer = func(val any) string {
	if d, ok := val.(*time.Time); ok {
		if d == nil {
			return ""
		}

		if todoitem.IsTBD(*d) {
			return "TBD"
		}

		return d.Format("2006-01-02")
	}

	return fmt.Sprintf("%v", val)
}

// sliceTransformer formats string slices for printing
var sliceTransformer = func(val any) string {
	if s, ok := val.([]string); ok {
		return strings.Join(s, ",")
	}

	return fmt.Sprintf("%v", val) // Fallback
}

// getTodoTransformer takes programOpts and returns a transformer for displaying todo statuses
func getTodoTransformer(opts programOpts) func(val any) string {
	todoTransformer := func(val any) string {
		if t, ok := val.(todoitem.TodoStatus); ok {
			if opts.useNerdfont {
				switch t {
				case todoitem.StatusNotTodo:
					return ""
				case todoitem.StatusOpen:
					return " "
				case todoitem.StatusDone:
					return "󰄬"
				case todoitem.StatusCanceled:
					return "󰜺"
				}
			} else {
				switch t {
				case todoitem.StatusNotTodo:
					return "E"
				case todoitem.StatusOpen:
					return " "
				case todoitem.StatusDone:
					return "x"
				case todoitem.StatusCanceled:
					return "-"
				}
			}
		}

		return fmt.Sprintf("%v", val) // Fallback
	}
	return todoTransformer
}

// calcTableSize calculates the size of the columns in the table based on parameters and the size of the terminal
func calcTableSize(opts programOpts) (tableSize, error) {
	terminalWidth, _, err := term.GetSize(0) // Get the current terminal size
	if err != nil {
		return tableSize{}, fmt.Errorf("error calculating tableSize: %v", err)
	}

	// Need enough space to give meaningful output
	if terminalWidth < minTotalWidth {
		return tableSize{}, fmt.Errorf("terminal too narrow for output (%d < %d)", terminalWidth, minTotalWidth)
	}

	// Use file width if needed
	numberOfCols := 5 // Number of columns
	var fileWidth int
	if opts.listFile {
		fileWidth = maxFileWidth
		numberOfCols++
	} else {
		fileWidth = 0
	}

	// Calculate the table sizes
	tz := tableSize{
		wMaxDate:        maxDateWidth,
		wMaxResponsible: maxResponisbleWidth,
		wMaxStatus:      maxStatusWidth,
		wMaxFile:        fileWidth,
	}

	renderOverhead := (numberOfCols + 1) + (numberOfCols-1)*2 // Rendering overhead. Separators + padding
	// Calculate the min width to use for the message column
	minMessageWidth := minTotalWidth - maxDateWidth - maxResponisbleWidth - maxStatusWidth - minTagsWidth - renderOverhead

	// The relative size of the dynamic columns to give to msg
	const ratioMsg = 0.8
	dynCols := terminalWidth - tz.wMaxDate - tz.wMaxResponsible - tz.wMaxStatus - tz.wMaxFile - renderOverhead // Available columns for dynamics sizing
	tz.wMinMsg = max(
		int(ratioMsg*float64(dynCols)), // Dynamic size
		minMessageWidth,
	)
	tz.wMinTags = max(
		dynCols-tz.wMinMsg, // Left from msg size
		minTagsWidth,       // Min allowed size
	)

	return tz, nil
}

// parseFlags parses the input flags and returns programOpts
func parseFlags() programOpts {
	opts := programOpts{}

	// Command line variables
	flag.StringVar(&opts.rootDir, "d", ".", "[D]irectory to parse")
	flag.BoolVar(&opts.recurse, "r", false, "[R]ecurse subdirectories")
	flag.StringVar(&opts.outputFormat, "f", outputFormats[0], fmt.Sprintf("Output [f]ormat %v", outputFormats))
	flag.BoolVar(&opts.useNerdfont, "u", true, "[U]se nerdfont symbols. May need to be false on older terminals")
	flag.StringVar(&opts.sortBy, "s", sortColumns[0], fmt.Sprintf("[S]ort by [%v]", sortColumns))
	flag.BoolVar(&opts.sortAsc, "a", true, "Sort [a]scending [true/false]")
	flag.BoolVar(&opts.listFile, "l", true, "[L]ist file location in table")
	flag.BoolVar(&opts.showVersion, "v", false, "Show current app version and exit")

	var filterInput string
	flag.StringVar(&filterInput, "i", "open,done,canceled", fmt.Sprintf("[I]nclude statuses. Multiple choices possible, delimit with ','. Allowed: %v", todoitem.GetAvailableStatuses()))
	flag.Parse()

	// Additional parsing of input
	opts.filterByStatus = strings.Split(filterInput, ",")

	return opts
}

// getTodos finds the files to check and calls parsing on them
func getTodos(dir string, recurse bool) ([]todoitem.Todo, error) {
	var todos []todoitem.Todo
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

// parseFile takes a path and a todo slice and adds todos from file. It returs a todoitem.Todo slice and and error
func parseFile(path string, todos []todoitem.Todo) ([]todoitem.Todo, error) {
	file, err := os.Open(path)
	if err != nil {
		return todos, fmt.Errorf("opening %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		todo, err := todoitem.ParseTodoLine(line, filepath.Base(path))
		if err != nil {
			continue
		}

		if todo.Status == todoitem.StatusNotTodo {
			continue
		}

		todos = append(todos, todo)
	}

	if err := scanner.Err(); err != nil {
		return todos, err
	}

	return todos, nil
}
