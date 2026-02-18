package lint

import "fmt"

// Severity represents the severity of a lint issue.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "Error"
	case SeverityWarning:
		return "Warning"
	case SeverityInfo:
		return "Info"
	default:
		return "Unknown"
	}
}

// Category represents the category of a lint check.
type Category string

const (
	CategorySpec  Category = "spec"
	CategoryStyle Category = "style"
)

// Severity returns the severity associated with the category.
func (c Category) Severity() Severity {
	switch c {
	case CategorySpec:
		return SeverityError
	case CategoryStyle:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// Issue represents a single linting issue found in a skill.
type Issue struct {
	Path     string
	Line     int
	Column   int
	Message  string
	Category Category
}

func (i Issue) String() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", i.Path, i.Line, i.Column, i.Category, i.Message)
}

// Formatter is the interface for formatting lint issues.
type Formatter interface {
	Format(issues []Issue) (string, error)
}
