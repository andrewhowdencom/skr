package lint

import (
	"fmt"
)

// Checker is the interface that all lint rules must implement.
type Checker interface {
	Name() string
	Check(doc *Document) ([]Issue, error)
}

// Fixer is an optional interface for rules that can autofix issues.
type Fixer interface {
	Checker
	Fix(doc *Document) error
}

type Linter struct {
	Checks []Checker
}

func NewLinter(checks []Checker) *Linter {
	return &Linter{
		Checks: checks,
	}
}

func (l *Linter) Run(dir string, fix bool) ([]Issue, error) {
	doc, err := NewDocument(dir)
	if err != nil {
		// If we cannot load the document (e.g., file missing or invalid YAML),
		// we treat it as a fatal error for now.
		return nil, err
	}

	var allIssues []Issue
	for _, check := range l.Checks {
		// Run Check
		issues, err := check.Check(doc)
		if err != nil {
			return nil, err
		}

		// If fix mode is enabled and issues were found, attempt to fix them.
		if fix && len(issues) > 0 {
			if fixer, ok := check.(Fixer); ok {
				if err := fixer.Fix(doc); err != nil {
					return nil, fmt.Errorf("failed to fix %s: %w", check.Name(), err)
				}

				// Mark issues as fixed.
				for i := range issues {
					issues[i].Fixed = true
				}

				// Flush any changes made by the Fixer to disk.
				if err := doc.Flush(); err != nil {
					return nil, err
				}
			}
		}

		allIssues = append(allIssues, issues...)
	}
	return allIssues, nil
}
