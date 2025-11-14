package parser

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	gast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"gantt-gen/model"
)

type tableContext struct {
	currentTask     *model.Task
	currentCalendar *model.Calendar
	tableType       string // "property", "dependency", "calendar"
	headers         []string
	rows            [][]string
}

// Add field to track context during parsing
type parseContext struct {
	project              *model.Project
	currentTaskIndex     int // Changed from *model.Task
	currentCalendarIndex int // Changed from *model.Calendar
	tableCtx             *tableContext
}

// Parse parses markdown and returns a Project
func Parse(source []byte) (*model.Project, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)
	doc := md.Parser().Parse(text.NewReader(source))

	ctx := &parseContext{
		project:              &model.Project{},
		currentTaskIndex:     -1,
		currentCalendarIndex: -1,
	}

	// Walk the AST
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			text := extractText(node, source)

			if node.Level == 1 {
				ctx.project.Name = text
				ctx.currentTaskIndex = -1
				ctx.currentCalendarIndex = -1
			} else if strings.HasPrefix(text, "Calendar:") {
				// Extract calendar name
				calName := strings.TrimSpace(strings.TrimPrefix(text, "Calendar:"))
				cal := model.Calendar{
					Name: calName,
				}
				ctx.project.Calendars = append(ctx.project.Calendars, cal)
				ctx.currentCalendarIndex = len(ctx.project.Calendars) - 1
				ctx.currentTaskIndex = -1
			} else {
				task := model.Task{
					Name:  text,
					Level: node.Level,
				}
				ctx.project.Tasks = append(ctx.project.Tasks, task)
				ctx.currentTaskIndex = len(ctx.project.Tasks) - 1
				ctx.currentCalendarIndex = -1
			}

		case *ast.Paragraph:
			// Check for milestones (bold text)
			if child := node.FirstChild(); child != nil {
				if emphasis, ok := child.(*ast.Emphasis); ok && emphasis.Level == 2 {
					text := extractText(emphasis, source)
					task := model.Task{
						Name:        text,
						IsMilestone: true,
						Level:       0,
					}
					ctx.project.Tasks = append(ctx.project.Tasks, task)
					ctx.currentTaskIndex = len(ctx.project.Tasks) - 1
					ctx.currentCalendarIndex = -1
				}
			}

		case *gast.Table:
			handleTable(node, source, ctx)
		}

		return ast.WalkContinue, nil
	})

	return ctx.project, nil
}

func (ctx *parseContext) currentTask() *model.Task {
	if ctx.currentTaskIndex >= 0 && ctx.currentTaskIndex < len(ctx.project.Tasks) {
		return &ctx.project.Tasks[ctx.currentTaskIndex]
	}
	return nil
}

func (ctx *parseContext) currentCalendar() *model.Calendar {
	if ctx.currentCalendarIndex >= 0 && ctx.currentCalendarIndex < len(ctx.project.Calendars) {
		return &ctx.project.Calendars[ctx.currentCalendarIndex]
	}
	return nil
}

func handleTable(table *gast.Table, source []byte, ctx *parseContext) {
	var headers []string
	var rows [][]string

	// Extract table data from TableHeader and TableRow nodes
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *gast.TableHeader:
			// Extract header cells directly (no TableRow wrapper)
			for cell := node.FirstChild(); cell != nil; cell = cell.NextSibling() {
				text := extractText(cell, source)
				headers = append(headers, strings.TrimSpace(text))
			}
		case *gast.TableRow:
			// Extract data row cells
			var cells []string
			for cell := node.FirstChild(); cell != nil; cell = cell.NextSibling() {
				text := extractText(cell, source)
				cells = append(cells, strings.TrimSpace(text))
			}
			rows = append(rows, cells)
		}
	}

	// Determine table type and process
	if len(headers) >= 2 {
		if headers[0] == "Property" && headers[1] == "Value" {
			parsePropertyTable(rows, ctx)
		} else if headers[0] == "Depends On" && headers[1] == "Type" {
			parseDependencyTable(rows, ctx)
		} else if headers[0] == "Type" && headers[1] == "Value" {
			parseCalendarTable(rows, ctx)
		}
	}
}

func parsePropertyTable(rows [][]string, ctx *parseContext) {
	task := ctx.currentTask()
	if task == nil {
		return
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		key := row[0]
		value := row[1]

		switch key {
		case "Start":
			if t, err := dateparse.ParseAny(value); err == nil {
				task.Start = &t
			}
		case "End":
			if t, err := dateparse.ParseAny(value); err == nil {
				task.End = &t
			}
		case "Date":
			if t, err := dateparse.ParseAny(value); err == nil {
				task.Date = &t
			}
		case "Duration":
			task.Duration = parseDuration(value)
		case "Link":
			task.Link = value
		case "Calendar":
			task.CalendarName = value
		}
	}
}

func parseDependencyTable(rows [][]string, ctx *parseContext) {
	task := ctx.currentTask()
	if task == nil {
		return
	}

	for _, row := range rows {
		if len(row) < 1 || row[0] == "-" {
			continue
		}

		depType := model.FinishToStart
		if len(row) >= 2 && row[1] != "" {
			depType = model.DependencyType(row[1])
		}

		dep := model.Dependency{
			TaskName: row[0],
			Type:     depType,
		}
		task.Dependencies = append(task.Dependencies, dep)
	}
}

func parseCalendarTable(rows [][]string, ctx *parseContext) {
	cal := ctx.currentCalendar()
	if cal == nil {
		return
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		key := row[0]
		value := row[1]

		switch key {
		case "Default":
			cal.IsDefault = strings.ToLower(value) == "true"
		case "Weekends":
			cal.Weekends = parseWeekends(value)
		case "Holiday":
			if t, err := dateparse.ParseAny(value); err == nil {
				cal.Holidays = append(cal.Holidays, t)
			}
		}
	}
}

func parseWeekends(s string) []time.Weekday {
	parts := strings.Split(s, ",")
	var weekends []time.Weekday

	dayMap := map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}

	for _, part := range parts {
		day := strings.ToLower(strings.TrimSpace(part))
		if len(day) > 3 {
			day = day[:3]
		}
		if wd, ok := dayMap[day]; ok {
			weekends = append(weekends, wd)
		}
	}

	return weekends
}

func parseDuration(s string) int {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	switch unit {
	case 'd':
		return num
	case 'w':
		return num * 7
	default:
		return 0
	}
}

func extractText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *ast.Text:
			buf.Write(node.Segment.Value(source))
		case *ast.AutoLink:
			buf.Write(node.URL(source))
		default:
			// Recursively extract text from other nodes
			if text := extractText(child, source); text != "" {
				buf.WriteString(text)
			}
		}
	}
	return buf.String()
}
