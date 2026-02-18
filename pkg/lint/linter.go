package lint

type Linter struct {
	Checks []Check
}

type Check interface {
	Check(dir string) ([]Issue, error)
}

func NewLinter(checks []Check) *Linter {
	return &Linter{
		Checks: checks,
	}
}

func (l *Linter) Run(dir string) ([]Issue, error) {
	var allIssues []Issue
	for _, check := range l.Checks {
		issues, err := check.Check(dir)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, issues...)
	}
	return allIssues, nil
}
