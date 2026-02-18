package lint

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
)

func TestFormatGNU(t *testing.T) {
	issues := []Issue{
		{
			Path:     "file.go",
			Line:     10,
			Column:   5,
			Message:  "error message",
			Category: CategorySpec,
		},
	}

	expected := "file.go:10:5: spec: error message\n"
	output, err := FormatGNU(issues)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestFormatSARIF(t *testing.T) {
	issues := []Issue{
		{
			Path:     "file.go",
			Line:     10,
			Column:   5,
			Message:  "error message",
			Category: CategorySpec,
		},
	}

	output, err := FormatSARIF(issues)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if len(log.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(log.Runs))
	}

	if len(log.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(log.Runs[0].Results))
	}

	result := log.Runs[0].Results[0]
	if result.RuleID != "spec" {
		t.Errorf("expected ruleId 'spec', got %q", result.RuleID)
	}
	if result.Level != "error" {
		t.Errorf("expected level 'error', got %q", result.Level)
	}
	if result.Message.Text != "error message" {
		t.Errorf("expected message 'error message', got %q", result.Message.Text)
	}
}

func TestFormatCheckstyle(t *testing.T) {
	issues := []Issue{
		{
			Path:     "file.go",
			Line:     10,
			Column:   5,
			Message:  "error message",
			Category: CategorySpec,
		},
	}

	output, err := FormatCheckstyle(issues)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(output, xml.Header) {
		t.Error("expected XML header")
	}

	if !strings.Contains(output, "<checkstyle") {
		t.Error("expected checkstyle root element")
	}

	if !strings.Contains(output, `<file name="file.go">`) {
		t.Error("expected file element")
	}

	// We expect severity="error" because CategorySpec maps to SeverityError
	if !strings.Contains(output, `severity="error"`) {
		t.Error("expected severity='error'")
	}
}
