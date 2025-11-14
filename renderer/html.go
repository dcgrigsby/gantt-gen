package renderer

import (
	"bytes"
	"fmt"
	"html/template"
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
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            padding: 20px;
            background: #f5f5f5;
        }

        .container {
            min-width: {{.MinWidth}}px;
            width: fit-content;
            max-width: 100%;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 20px;
        }

        .gantt-wrapper {
            overflow-x: auto;
            overflow-y: visible;
        }

        h1 {
            margin-bottom: 30px;
            color: #333;
        }

        .gantt {
            display: grid;
            grid-template-columns: 200px 1fr;
            gap: 0;
            border: 1px solid #ddd;
            overflow: visible;
        }

        .gantt-header {
            display: contents;
        }

        .gantt-header-task,
        .gantt-header-timeline {
            background: #f8f9fa;
            padding: 10px;
            font-weight: 600;
            border-bottom: 2px solid #ddd;
        }

        .task-row {
            display: contents;
        }

        .task-name {
            padding: 10px;
            border-bottom: 1px solid #eee;
            background: white;
            min-height: 40px;
        }

        .task-name.level-2 {
            font-weight: 600;
            padding-left: 10px;
        }

        .task-name.level-3 {
            padding-left: 30px;
        }

        .task-name.level-4 {
            padding-left: 50px;
        }

        .task-name.milestone {
            font-style: italic;
            color: #666;
            padding: 15px 10px;
        }

        .task-name a {
            color: #0066cc;
            text-decoration: none;
            font-size: 0.9em;
            margin-left: 5px;
        }

        .task-timeline {
            padding: 10px;
            border-bottom: 1px solid #eee;
            position: relative;
            background: white;
            overflow: visible;
            min-height: 40px;
        }

        .task-bar {
            position: absolute;
            height: 24px;
            top: 50%;
            transform: translateY(-50%);
            border-radius: 4px;
            background: #4a90e2;
            display: flex;
            align-items: center;
            padding: 0 8px;
            color: white;
            font-size: 0.85em;
            white-space: nowrap;
        }

        .task-bar.level-2 {
            background: #4a90e2;
        }

        .task-bar.level-3 {
            background: #7eb0e8;
        }

        .task-bar.level-4 {
            background: #a8c9ed;
        }

        .task-bar.milestone {
            width: 12px !important;
            height: 12px;
            border-radius: 50%;
            background: #e74c3c;
            transform: translateY(-50%) rotate(45deg);
            padding: 0;
        }

        .legend {
            margin-top: 20px;
            padding: 15px;
            background: #f8f9fa;
            border-radius: 4px;
        }

        .legend-item {
            display: inline-block;
            margin-right: 20px;
            font-size: 0.9em;
        }

        .legend-color {
            display: inline-block;
            width: 20px;
            height: 12px;
            margin-right: 5px;
            border-radius: 2px;
            vertical-align: middle;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Name}}</h1>

        <div class="gantt-wrapper">
            <div class="gantt">
                <div class="gantt-header">
                    <div class="gantt-header-task">Task</div>
                    <div class="gantt-header-timeline">Timeline</div>
                </div>

                {{range .Tasks}}
                <div class="task-row">
                    <div class="task-name {{if .IsMilestone}}milestone{{else}}level-{{.Level}}{{end}}">
                        {{.Name}}
                        {{if .Link}}<a href="{{.Link}}" target="_blank">🔗</a>{{end}}
                    </div>
                    <div class="task-timeline">
                        <div class="task-bar {{if .IsMilestone}}milestone{{else}}level-{{.Level}}{{end}}"
                             style="left: {{.BarLeft}}%; width: {{.BarWidth}}%;">
                            {{if not .IsMilestone}}{{.DateRange}}{{end}}
                        </div>
                    </div>
                </div>
                {{end}}
            </div>
        </div>

        <div class="legend">
            <div class="legend-item">
                <span class="legend-color" style="background: #4a90e2;"></span>
                H2 Tasks
            </div>
            <div class="legend-item">
                <span class="legend-color" style="background: #7eb0e8;"></span>
                H3 Tasks
            </div>
            <div class="legend-item">
                <span class="legend-color" style="background: #e74c3c; width: 12px; height: 12px; transform: rotate(45deg);"></span>
                Milestones
            </div>
        </div>
    </div>
</body>
</html>
`

type htmlTask struct {
	model.Task
	BarLeft   float64
	BarWidth  float64
	DateRange string
}

type htmlData struct {
	Name          string
	Tasks         []htmlTask
	MinWidth      int
}

// RenderHTML generates an HTML Gantt chart
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

	// Calculate dynamic width: ~25px per day, minimum 1200px
	// This ensures readability for both short and long projects
	taskColumnWidth := 220
	timelineMinWidth := 1000
	pixelsPerDay := 25.0
	calculatedTimelineWidth := int(totalDays * pixelsPerDay)
	if calculatedTimelineWidth < timelineMinWidth {
		calculatedTimelineWidth = timelineMinWidth
	}
	minWidth := taskColumnWidth + calculatedTimelineWidth

	// Build HTML tasks
	var htmlTasks []htmlTask
	for _, task := range project.Tasks {
		ht := htmlTask{Task: task}

		if task.CalculatedStart != nil && task.CalculatedEnd != nil {
			startOffset := task.CalculatedStart.Sub(minDate).Hours() / 24
			endOffset := task.CalculatedEnd.Sub(minDate).Hours() / 24

			ht.BarLeft = (startOffset / totalDays) * 100
			ht.BarWidth = ((endOffset - startOffset) / totalDays) * 100

			if ht.BarWidth < 0.5 {
				ht.BarWidth = 0.5 // Minimum width for visibility
			}

			ht.DateRange = fmt.Sprintf("%s - %s",
				task.CalculatedStart.Format("Jan 2"),
				task.CalculatedEnd.Format("Jan 2"))
		}

		htmlTasks = append(htmlTasks, ht)
	}

	data := htmlData{
		Name:     project.Name,
		Tasks:    htmlTasks,
		MinWidth: minWidth,
	}

	tmpl, err := template.New("gantt").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
