package renderer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"gantt-gen/model"
)

const confluenceTemplate = `<div style="display: flex; max-width: 100%; overflow-x: auto; border: 1px solid #ddd; background: white;">
    <!-- Fixed Task Column -->
    <div style="position: sticky; left: 0; z-index: 10; background: white; border-right: 2px solid #e0e0e0; flex-shrink: 0;">
        {{.TaskColumnSVG}}
    </div>

    <!-- Scrollable Timeline -->
    <div style="flex: 1; overflow-x: auto;">
        {{.TimelineSVG}}
    </div>
</div>

<!-- Usage Instructions -->
<div style="margin-top: 20px; padding: 15px; background: #f5f5f5; border-left: 4px solid #4a90e2; font-family: Arial, sans-serif; font-size: 14px;">
    <strong>To use in Confluence:</strong>
    <ol style="margin: 10px 0; padding-left: 20px;">
        <li>Copy all the HTML above (from &lt;div style="display: flex"...&gt; to &lt;/div&gt;)</li>
        <li>In Confluence, insert the <strong>HTML macro</strong></li>
        <li>Paste the HTML into the macro</li>
        <li>Save the page</li>
    </ol>
    <p style="margin: 10px 0 0 0; color: #666;">
        <em>Note: The task names column will stay fixed while you scroll the timeline horizontally.</em>
    </p>
</div>
`

// RenderConfluence generates a minimal HTML snippet for Confluence
func RenderConfluence(project *model.Project) (string, error) {
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

	// Generate task column (reuse from html.go)
	taskColumnSVG, err := renderTaskColumn(project, totalHeight)
	if err != nil {
		return "", err
	}

	// Generate timeline (reuse from html.go)
	timelineSVG, err := renderTimeline(project, minDate, maxDate, timelineWidth, effectiveTimelineWidth, totalHeight, milestonePadding)
	if err != nil {
		return "", err
	}

	// Combine into minimal Confluence HTML
	data := htmlData{
		Name:          project.Name,
		TaskColumnSVG: taskColumnSVG,
		TimelineSVG:   timelineSVG,
	}

	tmpl, err := template.New("confluence").Parse(confluenceTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
