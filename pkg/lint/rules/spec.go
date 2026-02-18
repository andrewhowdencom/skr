package rules

import (
	"regexp"

	"github.com/andrewhowdencom/skr/pkg/lint"
	"gopkg.in/yaml.v3"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
)

type Name struct{}

func (r *Name) Name() string {
	return "spec/name"
}

func (r *Name) Check(doc *lint.Document) ([]lint.Issue, error) {
	var issues []lint.Issue

	if doc.Skill == nil {
		return nil, nil // Cannot check name if skill not parsed
	}

	name := doc.Skill.Name
	// Find node for accurate line number reporting.
	var line, col int
	node := findKeyNode(doc.Tree, "name")
	if node != nil {
		line = node.Line
		// The value node contains the actual content and location of the value.
		valNode := findValueNode(doc.Tree, "name")
		if valNode != nil {
			line = valNode.Line
			col = valNode.Column
		}
	}

	if name == "" {
		issues = append(issues, lint.Issue{
			Path:     doc.Path,
			Line:     line,
			Column:   col,
			Message:  "name is required",
			Category: lint.CategorySpec,
		})
	} else {
		if len(name) > 64 {
			issues = append(issues, lint.Issue{
				Path:     doc.Path,
				Line:     line,
				Column:   col,
				Message:  "name must be 64 characters or less",
				Category: lint.CategorySpec,
			})
		}
		if !validNameRegex.MatchString(name) {
			issues = append(issues, lint.Issue{
				Path:     doc.Path,
				Line:     line,
				Column:   col,
				Message:  "name must contain only lowercase alphanumeric characters and hyphens",
				Category: lint.CategorySpec,
			})
		}
	}

	return issues, nil
}

type Description struct{}

func (r *Description) Name() string {
	return "spec/description"
}

func (r *Description) Check(doc *lint.Document) ([]lint.Issue, error) {
	var issues []lint.Issue
	if doc.Skill == nil {
		return nil, nil
	}

	desc := doc.Skill.Description
	var line, col int
	valNode := findValueNode(doc.Tree, "description")
	if valNode != nil {
		line = valNode.Line
		col = valNode.Column
	}

	if desc == "" {
		issues = append(issues, lint.Issue{
			Path:     doc.Path,
			Line:     line,
			Column:   col,
			Message:  "description is required",
			Category: lint.CategorySpec,
		})
	} else {
		if len(desc) > 1024 {
			issues = append(issues, lint.Issue{
				Path:     doc.Path,
				Line:     line,
				Column:   col,
				Message:  "description must be 1024 characters or less",
				Category: lint.CategorySpec,
			})
		}
	}

	return issues, nil
}

// Helpers

// findValueNode locates the value node for a given key in the YAML document.
// It assumes the root is a DocumentNode containing a MappingNode.
func findValueNode(root *yaml.Node, key string) *yaml.Node {
	if root == nil {
		return nil
	}
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		mapping := root.Content[0]
		if mapping.Kind == yaml.MappingNode {
			for i := 0; i < len(mapping.Content); i += 2 {
				if mapping.Content[i].Value == key {
					// Check bounds
					if i+1 < len(mapping.Content) {
						return mapping.Content[i+1]
					}
				}
			}
		}
	}
	return nil
}

func findKeyNode(root *yaml.Node, key string) *yaml.Node {
	if root == nil {
		return nil
	}
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		mapping := root.Content[0]
		if mapping.Kind == yaml.MappingNode {
			for i := 0; i < len(mapping.Content); i += 2 {
				if mapping.Content[i].Value == key {
					return mapping.Content[i]
				}
			}
		}
	}
	return nil
}
