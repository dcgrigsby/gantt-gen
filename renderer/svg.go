package renderer

import (
	"bytes"
	"fmt"
	"strings"
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
    <text x="30" y="70" font-family="Arial, sans-serif" font-size="14" font-weight="600" fill="#333">
        Task
    </text>

    <text x="230" y="70" font-family="Arial, sans-serif" font-size="14" font-weight="600" fill="#333">
        Timeline
    </text>

    <!-- Tasks -->
    {{range $i, $task := .Tasks}}
    <g class="task-row">
        <!-- Task name -->
        <text x="{{$task.NameX}}" y="{{$task.TextY}}"
              font-family="Arial, sans-serif" font-size="13"
              {{if $task.IsMilestone}}font-style="italic" fill="#666"{{else}}fill="#333"{{end}}>
            {{$task.DisplayName}}
        </text>

        <!-- Timeline -->

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
    </g>
    {{end}}
</svg>
`

const (
	maxTaskNameWidth   = 170 // Reserve 30px for indentation/padding
	avgCharWidthPixels = 7.0 // Average character width in Arial 13px
	ellipsis           = "..."
)

// truncateTaskName truncates task name to fit within maxWidth pixels
func truncateTaskName(name string, level int) string {
	// Calculate indent space used
	indent := 0
	if level >= 2 {
		indent = (level - 2) * 20
	}

	availableWidth := maxTaskNameWidth - indent
	maxChars := int(float64(availableWidth) / avgCharWidthPixels)

	// Account for multi-byte UTF-8 characters (rough estimate)
	runeCount := len([]rune(name))
	if runeCount <= maxChars {
		return name
	}

	// Truncate and add ellipsis
	runes := []rune(name)
	if maxChars > len(ellipsis) {
		truncated := string(runes[:maxChars-len(ellipsis)])
		// Trim trailing spaces before adding ellipsis
		truncated = strings.TrimRight(truncated, " ")
		return truncated + ellipsis
	}

	return ellipsis
}

type svgTask struct {
	model.Task
	DisplayName      string  // Truncated name for display
	Y                int
	NameX            int
	TextY            int
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
}

type svgData struct {
	Name          string
	Width         int
	Height        int
	TimelineWidth int
	Tasks         []svgTask
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

	// Reserve space for milestone diamonds at the edges
	// A rotated 10x10 square has diagonal = 10*sqrt(2) ≈ 14.14px
	// Radius from center to corner = 14.14/2 ≈ 7.07px
	const milestoneRadius = 7.07
	milestonePadding := milestoneRadius * 2 // Total padding needed
	effectiveTimelineWidth := float64(timelineWidth) - milestonePadding

	rowHeight := 40
	headerHeight := 80

	// Build SVG tasks
	var svgTasks []svgTask
	for i, task := range project.Tasks {
		y := headerHeight + (i * rowHeight)

		// Truncate name for display
		displayName := truncateTaskName(task.Name, task.Level)

		st := svgTask{
			Task:        task,
			DisplayName: displayName,
			Y:           y,
			NameX:       30 + (task.Level-2)*20,
			TextY:       y + 25,
		}

		if task.IsMilestone {
			st.NameX = 30
			displayName = truncateTaskName(task.Name, 0)
			st.DisplayName = displayName
		}

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

			// Use effective width for positioning to reserve space for milestones
			barLeft := (startOffset / totalDays) * effectiveTimelineWidth
			barWidth := ((endOffset - startOffset) / totalDays) * effectiveTimelineWidth

			if barWidth < 5 {
				barWidth = 5 // Minimum width for visibility
			}

			// Add left padding to keep diamonds within timeline bounds
			// Diamond extends (milestoneRadius - 5) pixels left of BarX
			leftPadding := milestoneRadius - 5.0 // 7.07 - 5 = 2.07
			st.BarX = 220 + barLeft + leftPadding
			st.BarY = y + 6
			st.BarWidth = barWidth
			st.MilestoneY = y + 16
			st.MilestoneCenterX = st.BarX + 5  // Center X of 10px square
			st.MilestoneCenterY = st.MilestoneY + 5  // Center Y of 10px square
			st.DateX = st.BarX + 5
			st.DateY = y + 24

			st.DateRange = fmt.Sprintf("%s - %s",
				task.CalculatedStart.Format("Jan 2"),
				task.CalculatedEnd.Format("Jan 2"))
		}

		svgTasks = append(svgTasks, st)
	}

	totalHeight := headerHeight + (len(project.Tasks) * rowHeight) + 20 // just add minimal bottom padding
	totalWidth := 220 + timelineWidth + 20 // task column + timeline + minimal right padding

	data := svgData{
		Name:          project.Name,
		Width:         totalWidth,
		Height:        totalHeight,
		TimelineWidth: timelineWidth,
		Tasks:         svgTasks,
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
