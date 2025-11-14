package parser

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"gantt-gen/model"
)

// Parse parses markdown and returns a Project
func Parse(source []byte) (*model.Project, error) {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))

	project := &model.Project{}

	// Walk the AST
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			text := extractText(node, source)

			if node.Level == 1 {
				project.Name = text
			} else {
				task := model.Task{
					Name:  text,
					Level: node.Level,
				}
				project.Tasks = append(project.Tasks, task)
			}

		case *ast.Paragraph:
			// Check if paragraph contains strong emphasis (bold)
			if child := node.FirstChild(); child != nil {
				if emphasis, ok := child.(*ast.Emphasis); ok && emphasis.Level == 2 {
					text := extractText(emphasis, source)
					task := model.Task{
						Name:        text,
						IsMilestone: true,
						Level:       0,
					}
					project.Tasks = append(project.Tasks, task)
				}
			}
		}

		return ast.WalkContinue, nil
	})

	return project, nil
}

func extractText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			buf.Write(text.Segment.Value(source))
		}
	}
	return buf.String()
}
