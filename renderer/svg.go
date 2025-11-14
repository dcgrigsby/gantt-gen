package renderer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"gantt-gen/model"
)

const svgTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="{{.Width}}" height="{{.Height}}" viewBox="0 0 {{.Width}} {{.Height}}">
    <!-- Background -->
    <rect width="{{.Width}}" height="{{.Height}}" fill="#ffffff"/>

    <!-- Title -->
    <text x="20" y="30" font-family="Arial, sans-serif" font-size="20" font-weight="bold" fill="#333">
        {{.Name}}
    </text>

    <!-- Column headers -->
    <rect x="20" y="50" width="200" height="30" fill="#f8f9fa" stroke="#ddd"/>
    <text x="30" y="70" font-family="Arial, sans-serif" font-size="14" font-weight="600" fill="#333">
        Task
    </text>

    <rect x="220" y="50" width="{{.TimelineWidth}}" height="30" fill="#f8f9fa" stroke="#ddd"/>
    <text x="230" y="70" font-family="Arial, sans-serif" font-size="14" font-weight="600" fill="#333">
        Timeline
    </text>

    <!-- Tasks -->
    {{range $i, $task := .Tasks}}
    <g class="task-row">
        <!-- Task name -->
        <rect x="20" y="{{$task.Y}}" width="200" height="30" fill="white" stroke="#eee"/>
        <text x="{{$task.NameX}}" y="{{$task.TextY}}"
              font-family="Arial, sans-serif" font-size="13"
              {{if $task.IsMilestone}}font-style="italic" fill="#666"{{else}}fill="#333"{{end}}>
            {{$task.Name}}
        </text>
        {{if $task.Link}}
        <text x="{{$task.LinkX}}" y="{{$task.TextY}}" font-size="11" fill="#0066cc">🔗</text>
        {{end}}

        <!-- Timeline -->
        <rect x="220" y="{{$task.Y}}" width="{{$.TimelineWidth}}" height="30" fill="white" stroke="#eee"/>

        <!-- Task bar or milestone -->
        {{if $task.IsMilestone}}
        <rect x="{{$task.BarX}}" y="{{$task.MilestoneY}}" width="10" height="10"
              fill="#e74c3c" transform="rotate(45 {{$task.BarX}} {{$task.MilestoneY}})"/>
        {{else}}
        <rect x="{{$task.BarX}}" y="{{$task.BarY}}" width="{{$task.BarWidth}}" height="20"
              fill="{{$task.Color}}" rx="3"/>
        <text x="{{$task.DateX}}" y="{{$task.DateY}}"
              font-family="Arial, sans-serif" font-size="10" fill="white">
            {{$task.DateRange}}
        </text>
        {{end}}
    </g>
    {{end}}

    <!-- Legend -->
    <g class="legend" transform="translate(20, {{.LegendY}})">
        <rect width="{{.Width}}" height="60" fill="#f8f9fa" rx="4"/>

        <rect x="10" y="15" width="20" height="12" fill="#4a90e2" rx="2"/>
        <text x="35" y="25" font-family="Arial, sans-serif" font-size="12" fill="#333">H2 Tasks</text>

        <rect x="120" y="15" width="20" height="12" fill="#7eb0e8" rx="2"/>
        <text x="145" y="25" font-family="Arial, sans-serif" font-size="12" fill="#333">H3 Tasks</text>

        <rect x="230" y="15" width="12" height="12" fill="#e74c3c" transform="rotate(45 236 21)"/>
        <text x="255" y="25" font-family="Arial, sans-serif" font-size="12" fill="#333">Milestones</text>
    </g>
</svg>
`

type svgTask struct {
	model.Task
	Y           int
	NameX       int
	TextY       int
	LinkX       int
	BarX        float64
	BarY        int
	BarWidth    float64
	MilestoneY  int
	DateX       float64
	DateY       int
	DateRange   string
	Color       string
}

type svgData struct {
	Name          string
	Width         int
	Height        int
	TimelineWidth int
	Tasks         []svgTask
	LegendY       int
}

// RenderSVG generates an SVG Gantt chart
func RenderSVG(project *model.Project) (string, error) {
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

	// Calculate dynamic width: ~25px per day, minimum 800px
	pixelsPerDay := 25.0
	timelineWidth := int(totalDays * pixelsPerDay)
	if timelineWidth < 800 {
		timelineWidth = 800
	}

	rowHeight := 30
	headerHeight := 80

	// Build SVG tasks
	var svgTasks []svgTask
	for i, task := range project.Tasks {
		y := headerHeight + (i * rowHeight)

		st := svgTask{
			Task:  task,
			Y:     y,
			NameX: 30 + (task.Level-2)*20,
			TextY: y + 20,
		}

		if task.IsMilestone {
			st.NameX = 30
		}

		st.LinkX = st.NameX + len(task.Name)*7 + 5

		// Color by level
		switch task.Level {
		case 2:
			st.Color = "#4a90e2"
		case 3:
			st.Color = "#7eb0e8"
		case 4:
			st.Color = "#a8c9ed"
		default:
			st.Color = "#4a90e2"
		}

		if task.CalculatedStart != nil && task.CalculatedEnd != nil {
			startOffset := task.CalculatedStart.Sub(minDate).Hours() / 24
			endOffset := task.CalculatedEnd.Sub(minDate).Hours() / 24

			barLeft := (startOffset / totalDays) * float64(timelineWidth)
			barWidth := ((endOffset - startOffset) / totalDays) * float64(timelineWidth)

			if barWidth < 5 {
				barWidth = 5 // Minimum width for visibility
			}

			st.BarX = 220 + barLeft
			st.BarY = y + 5
			st.BarWidth = barWidth
			st.MilestoneY = y + 15
			st.DateX = st.BarX + 5
			st.DateY = y + 18

			st.DateRange = fmt.Sprintf("%s - %s",
				task.CalculatedStart.Format("Jan 2"),
				task.CalculatedEnd.Format("Jan 2"))
		}

		svgTasks = append(svgTasks, st)
	}

	totalHeight := headerHeight + (len(project.Tasks) * rowHeight) + 100
	totalWidth := 220 + timelineWidth + 100 // task column + timeline + padding

	data := svgData{
		Name:          project.Name,
		Width:         totalWidth,
		Height:        totalHeight,
		TimelineWidth: timelineWidth,
		Tasks:         svgTasks,
		LegendY:       totalHeight - 80,
	}

	tmpl, err := template.New("gantt").Parse(svgTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
