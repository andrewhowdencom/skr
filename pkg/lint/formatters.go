package lint

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// FormatGNU formats issues in the GNU error format.
// path/to/file:line:col: category: message
func FormatGNU(issues []Issue) (string, error) {
	var sb strings.Builder
	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("%s:%d:%d: %s: %s\n", issue.Path, issue.Line, issue.Column, issue.Category, issue.Message))
	}
	return sb.String(), nil
}

// SARIF structures (simplified)
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

type sarifLog struct {
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

// FormatSARIF formats issues in the SARIF format.
func FormatSARIF(issues []Issue) (string, error) {
	run := sarifRun{
		Tool: sarifTool{
			Driver: sarifDriver{
				Name:    "skr",
				Version: "1.0.0", // TODO: Get actual version
			},
		},
		Results: []sarifResult{},
	}

	for _, issue := range issues {
		level := "note"
		if issue.Category.Severity() == SeverityError {
			level = "error"
		} else if issue.Category.Severity() == SeverityWarning {
			level = "warning"
		}

		result := sarifResult{
			RuleID: string(issue.Category),
			Level:  level,
			Message: sarifMessage{
				Text: issue.Message,
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: issue.Path,
						},
						Region: sarifRegion{
							StartLine:   issue.Line,
							StartColumn: issue.Column,
						},
					},
				},
			},
		}
		run.Results = append(run.Results, result)
	}

	log := sarifLog{
		Version: "2.1.0",
		Runs:    []sarifRun{run},
	}

	bytes, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Checkstyle structures
type checkstyleFile struct {
	Name   string            `xml:"name,attr"`
	Errors []checkstyleError `xml:"error"`
}

type checkstyleError struct {
	Line     int    `xml:"line,attr"`
	Column   int    `xml:"column,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
}

type checkstyleOutput struct {
	XMLName xml.Name         `xml:"checkstyle"`
	Version string           `xml:"version,attr"`
	Files   []checkstyleFile `xml:"file"`
}

// FormatCheckstyle formats issues in the Checkstyle XML format.
func FormatCheckstyle(issues []Issue) (string, error) {
	// Group issues by file
	issuesByFile := make(map[string][]Issue)
	for _, issue := range issues {
		issuesByFile[issue.Path] = append(issuesByFile[issue.Path], issue)
	}

	var files []checkstyleFile
	for path, fileIssues := range issuesByFile {
		var errors []checkstyleError
		for _, issue := range fileIssues {
			severity := "info"
			if issue.Category.Severity() == SeverityError {
				severity = "error"
			} else if issue.Category.Severity() == SeverityWarning {
				severity = "warning"
			}

			errors = append(errors, checkstyleError{
				Line:     issue.Line,
				Column:   issue.Column,
				Severity: severity,
				Message:  issue.Message,
				Source:   "skr.lint." + string(issue.Category),
			})
		}
		files = append(files, checkstyleFile{
			Name:   path,
			Errors: errors,
		})
	}

	output := checkstyleOutput{
		Version: "8.0",
		Files:   files,
	}

	bytes, err := xml.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(bytes), nil
}
