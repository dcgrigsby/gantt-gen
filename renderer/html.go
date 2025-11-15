package renderer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"gantt-gen/model"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Name}} - Gantt Chart</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: Arial, sans-serif;
            background: #f5f5f5;
            overflow: hidden;
        }

        .container {
            display: flex;
            flex-direction: column;
            height: 100vh;
        }

        .header {
            background: white;
            padding: 20px;
            border-bottom: 2px solid #e0e0e0;
            flex-shrink: 0;
        }

        .header h1 {
            font-size: 24px;
            color: #333;
            margin: 0;
        }

        .gantt-container {
            display: flex;
            flex: 1;
            overflow: hidden;
            background: white;
        }

        .task-column {
            width: 240px;
            flex-shrink: 0;
            overflow-y: auto;
            border-right: 2px solid #e0e0e0;
            background: white;
        }

        .timeline-column {
            flex: 1;
            overflow: auto;
            background: white;
        }

        /* Synchronize scrolling between task and timeline vertically */
        .task-column,
        .timeline-column {
            scroll-behavior: smooth;
        }

        /* Custom scrollbar styling */
        .timeline-column::-webkit-scrollbar {
            width: 12px;
            height: 12px;
        }

        .timeline-column::-webkit-scrollbar-track {
            background: #f1f1f1;
        }

        .timeline-column::-webkit-scrollbar-thumb {
            background: #888;
            border-radius: 6px;
        }

        .timeline-column::-webkit-scrollbar-thumb:hover {
            background: #555;
        }

        svg {
            display: block;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Name}}</h1>
        </div>
        <div class="gantt-container">
            <div class="task-column" id="taskColumn">
                {{.TaskColumnSVG}}
            </div>
            <div class="timeline-column" id="timelineColumn">
                {{.TimelineSVG}}
            </div>
        </div>
    </div>

    <script>
        // Synchronize vertical scrolling between task column and timeline
        const taskColumn = document.getElementById('taskColumn');
        const timelineColumn = document.getElementById('timelineColumn');

        timelineColumn.addEventListener('scroll', () => {
            taskColumn.scrollTop = timelineColumn.scrollTop;
        });

        taskColumn.addEventListener('scroll', () => {
            timelineColumn.scrollTop = taskColumn.scrollTop;
        });
    </script>
</body>
</html>
`

const taskColumnSVGTemplate = `<svg xmlns="http://www.w3.org/2000/svg" width="240" height="{{.Height}}" viewBox="0 0 240 {{.Height}}">
    <rect width="240" height="{{.Height}}" fill="#ffffff"/>

    <!-- Column header -->
    <rect x="0" y="0" width="240" height="40" fill="#f8f9fa" stroke="#eee"/>
    <text x="10" y="25" font-family="Arial, sans-serif" font-size="13" font-weight="600" fill="#333">
        Tasks
    </text>

    <!-- Task rows -->
    {{range $task := .Tasks}}
    <rect x="0" y="{{$task.Y}}" width="240" height="40" fill="none" stroke="#eee"/>
    <text x="{{$task.NameX}}" y="{{$task.TextY}}"
          font-family="Arial, sans-serif" font-size="13"
          {{if $task.IsMilestone}}font-style="italic" fill="#666"{{else}}fill="#333"{{end}}>
        {{$task.DisplayName}}
    </text>
    {{end}}
</svg>`

const timelineSVGTemplate = `<svg xmlns="http://www.w3.org/2000/svg" width="{{.Width}}" height="{{.Height}}" viewBox="0 0 {{.Width}} {{.Height}}">
    <rect width="{{.Width}}" height="{{.Height}}" fill="#ffffff"/>

    <!-- Timeline header cells -->
    {{range $cell := .TimelineCells}}
    <rect x="{{$cell.X}}" y="0" width="{{$cell.Width}}" height="40" fill="#f8f9fa" stroke="#eee"/>
    <text x="{{$cell.TextX}}" y="25" font-family="Arial, sans-serif" font-size="12" font-weight="600" fill="#333" text-anchor="start">
        {{$cell.Label}}
    </text>
    {{end}}

    <!-- Task timeline rows -->
    {{range $task := .Tasks}}
    <rect x="0" y="{{$task.Y}}" width="{{$.Width}}" height="40" fill="none" stroke="#eee"/>

    <!-- Task bar or milestone -->
    {{if $task.IsMilestone}}
    <rect x="{{$task.BarX}}" y="{{$task.MilestoneY}}" width="10" height="10"
          fill="#e74c3c" transform="rotate(45 {{$task.MilestoneCenterX}} {{$task.MilestoneCenterY}})"/>
    {{else}}
    <rect x="{{$task.BarX}}" y="{{$task.BarY}}" width="{{$task.BarWidth}}" height="28"
          fill="{{$task.Color}}" rx="3"/>
    <text x="{{$task.DateX}}" y="{{$task.DateY}}"
          font-family="Arial, sans-serif" font-size="10" fill="white">
        {{$task.DateRange}}
    </text>
    {{end}}
    {{end}}
</svg>`

type htmlData struct {
	Name          string
	TaskColumnSVG string
	TimelineSVG   string
}

type taskColumnData struct {
	Height int
	Tasks  []taskColumnTask
}

type taskColumnTask struct {
	Y           int
	NameX       int
	TextY       int
	DisplayName string
	IsMilestone bool
}

type timelineData struct {
	Width         int
	Height        int
	TimelineCells []timelineHeaderCell
	Tasks         []timelineTask
}

type timelineHeaderCell struct {
	X     float64
	Width float64
	TextX float64
	Label string
}

type timelineTask struct {
	Y                int
	BarX             float64
	BarY             int
	BarWidth         float64
	MilestoneY       int
	MilestoneCenterX float64
	MilestoneCenterY int
	DateX            float64
	DateY            int
	DateRange        string
	Color            string
	IsMilestone      bool
}

// RenderHTML generates an HTML file with scrollable Gantt chart
func RenderHTML(project *model.Project) (string, error) {
	// Find date range
	var minDate, maxDate time.Time
	for _, task := range project.Tasks {
		if task.CalculatedStart != nil {
			if minDate.IsZero() || task.CalculatedStart.Before(minDate) {
				minDate = *task.CalculatedStart
			}
		}
		if task.CalculatedEnd != nil {
			if maxDate.IsZero() || task.CalculatedEnd.After(maxDate) {
				maxDate = *task.CalculatedEnd
			}
		}
	}

	if minDate.IsZero() || maxDate.IsZero() {
		return "", fmt.Errorf("no tasks with calculated dates")
	}

	totalDays := maxDate.Sub(minDate).Hours() / 24

	// Calculate timeline width: ~25px per day, minimum 800px
	pixelsPerDay := 25.0
	timelineWidth := int(totalDays * pixelsPerDay)
	if timelineWidth < 800 {
		timelineWidth = 800
	}

	// Reserve space for milestone diamonds
	const milestoneRadius = 7.07
	milestonePadding := milestoneRadius * 2
	effectiveTimelineWidth := float64(timelineWidth) - milestonePadding

	rowHeight := 40
	headerHeight := 40
	totalHeight := headerHeight + (len(project.Tasks) * rowHeight)

	// Generate task column
	taskColumnSVG, err := renderTaskColumn(project, totalHeight)
	if err != nil {
		return "", err
	}

	// Generate timeline
	timelineSVG, err := renderTimeline(project, minDate, maxDate, timelineWidth, effectiveTimelineWidth, totalHeight, milestonePadding)
	if err != nil {
		return "", err
	}

	// Combine into HTML
	data := htmlData{
		Name:          project.Name,
		TaskColumnSVG: taskColumnSVG,
		TimelineSVG:   timelineSVG,
	}

	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func renderTaskColumn(project *model.Project, height int) (string, error) {
	var tasks []taskColumnTask

	for i, task := range project.Tasks {
		y := 40 + (i * 40) // header height + row offset

		displayName := truncateTaskName(task.Name, task.Level)
		nameX := 10 + (task.Level-2)*20

		if task.IsMilestone {
			nameX = 10
			displayName = truncateTaskName(task.Name, 0)
		}

		tasks = append(tasks, taskColumnTask{
			Y:           y,
			NameX:       nameX,
			TextY:       y + 25,
			DisplayName: displayName,
			IsMilestone: task.IsMilestone,
		})
	}

	data := taskColumnData{
		Height: height,
		Tasks:  tasks,
	}

	tmpl, err := template.New("taskColumn").Parse(taskColumnSVGTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func renderTimeline(project *model.Project, minDate, maxDate time.Time, timelineWidth int, effectiveTimelineWidth float64, height int, milestonePadding float64) (string, error) {
	totalDays := maxDate.Sub(minDate).Hours() / 24

	// Add right padding to prevent milestone truncation
	const rightPadding = 20
	timelineSVGWidth := timelineWidth + rightPadding

	// Generate timeline header cells (use full width including padding)
	var headerCells []timelineHeaderCell
	cells := generateTimelineHeaderCells(minDate, maxDate, timelineSVGWidth)
	for _, cell := range cells {
		headerCells = append(headerCells, timelineHeaderCell{
			X:     cell.X,
			Width: cell.Width,
			TextX: cell.X + 5, // Add padding for text
			Label: cell.Label,
		})
	}

	// Generate task timeline rows
	var tasks []timelineTask
	for i, task := range project.Tasks {
		y := 40 + (i * 40) // header height + row offset

		tt := timelineTask{
			Y:           y,
			IsMilestone: task.IsMilestone,
		}

		// Color by level
		switch task.Level {
		case 2:
			tt.Color = "#4a90e2"
		case 3:
			tt.Color = "#7eb0e8"
		case 4:
			tt.Color = "#a8c9ed"
		default:
			tt.Color = "#4a90e2"
		}

		if task.CalculatedStart != nil && task.CalculatedEnd != nil {
			startOffset := task.CalculatedStart.Sub(minDate).Hours() / 24
			endOffset := task.CalculatedEnd.Sub(minDate).Hours() / 24

			barLeft := (startOffset / totalDays) * effectiveTimelineWidth
			barWidth := ((endOffset - startOffset) / totalDays) * effectiveTimelineWidth

			if barWidth < 5 {
				barWidth = 5
			}

			leftPadding := milestonePadding / 2
			tt.BarX = barLeft + leftPadding
			tt.BarY = y + 6
			tt.BarWidth = barWidth
			tt.MilestoneY = y + 16
			tt.MilestoneCenterX = tt.BarX + 5
			tt.MilestoneCenterY = tt.MilestoneY + 5
			tt.DateX = tt.BarX + 5
			tt.DateY = y + 24

			tt.DateRange = fmt.Sprintf("%s - %s",
				task.CalculatedStart.Format("Jan 2"),
				task.CalculatedEnd.Format("Jan 2"))
		}

		tasks = append(tasks, tt)
	}

	data := timelineData{
		Width:         timelineSVGWidth,
		Height:        height,
		TimelineCells: headerCells,
		Tasks:         tasks,
	}

	tmpl, err := template.New("timeline").Parse(timelineSVGTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateTimelineHeaderCells creates timeline header cells for HTML format
func generateTimelineHeaderCells(minDate, maxDate time.Time, timelineWidth int) []timelineCell {
	totalDays := maxDate.Sub(minDate).Hours() / 24
	fullWidth := float64(timelineWidth)

	var cells []timelineCell

	// Use months if > 60 days, otherwise use weeks
	if totalDays > 60 {
		// Generate month cells
		current := time.Date(minDate.Year(), minDate.Month(), 1, 0, 0, 0, 0, minDate.Location())

		for {
			monthStart := current
			monthEnd := monthStart.AddDate(0, 1, 0)

			if monthStart.After(maxDate) {
				break
			}

			visibleStart := monthStart
			if visibleStart.Before(minDate) {
				visibleStart = minDate
			}
			visibleEnd := monthEnd
			if visibleEnd.After(maxDate) {
				visibleEnd = maxDate
			}

			startOffset := visibleStart.Sub(minDate).Hours() / 24
			endOffset := visibleEnd.Sub(minDate).Hours() / 24

			x := (startOffset / totalDays) * fullWidth
			width := ((endOffset - startOffset) / totalDays) * fullWidth

			cells = append(cells, timelineCell{
				X:     x,
				Width: width,
				Label: monthStart.Format("Jan 2006"),
			})

			current = monthEnd
		}
	} else {
		// Generate week cells
		current := minDate
		weekday := int(current.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		current = current.AddDate(0, 0, 1-weekday)

		for {
			weekStart := current
			weekEnd := current.AddDate(0, 0, 7)

			if weekStart.After(maxDate) {
				break
			}

			visibleStart := weekStart
			if visibleStart.Before(minDate) {
				visibleStart = minDate
			}
			visibleEnd := weekEnd
			if visibleEnd.After(maxDate) {
				visibleEnd = maxDate
			}

			startOffset := visibleStart.Sub(minDate).Hours() / 24
			endOffset := visibleEnd.Sub(minDate).Hours() / 24

			x := (startOffset / totalDays) * fullWidth
			width := ((endOffset - startOffset) / totalDays) * fullWidth

			displayEnd := weekEnd.AddDate(0, 0, -1)
			if displayEnd.After(maxDate) {
				displayEnd = maxDate
			}

			cells = append(cells, timelineCell{
				X:     x,
				Width: width,
				Label: fmt.Sprintf("%s - %s", weekStart.Format("Jan 2"), displayEnd.Format("Jan 2")),
			})

			current = weekEnd
		}
	}

	return cells
}
