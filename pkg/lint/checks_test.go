package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewhowdencom/skr/pkg/skill"
)

func TestSpecCheck(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skr-test-spec")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test missing file
	check := SpecCheck{}
	issues, err := check.Check(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].Category != CategorySpec {
		t.Errorf("expected 1 spec issue for missing file, got %v", issues)
	}

	// Test invalid frontmatter
	skillPath := filepath.Join(tempDir, skill.SkillFileName)
	if err := os.WriteFile(skillPath, []byte("invalid content"), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	issues, err = check.Check(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) == 0 {
		t.Error("expected spec issue for invalid frontmatter")
	}

	// Test valid skill
	validContent := `---
name: valid-skill
description: A valid skill.
---
`
	if err := os.WriteFile(skillPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	issues, err = check.Check(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues for valid skill, got %v", issues)
	}

	// Test Invalid Name (Caps) - Check Line Number
	invalidNameContent := `---
name: Invalid-Skill
description: A valid description.
---
`
	if err := os.WriteFile(skillPath, []byte(invalidNameContent), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	issues, err = check.Check(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	// "name: Invalid-Skill" is on line 2 (offset by frontmatter start? no wait)
	// file:
	// 1: ---
	// 2: name: Invalid-Skill
	// 3: description: ...
	// 4: ---
	// yaml.Node for "Invalid-Skill" should report line 2.
	// My logic added +1 to line?
	// yaml.Node.Line for "name" key is 2. Value is 2.
	// checks.go: line := nameNode.Line + 1.
	// Wait, yaml.Node.Line is 1-based relative to the input string.
	// Input string is content[4 : 4+end].
	// Content starts at file index 0.
	// content[0:3] is "---\n".
	// So frontmatter slice starts at index 4, which is line 2 of the file.
	// If name is on line 1 of the frontmatter slice, it's line 2 of the file.
	// checks.go adds +1.
	// If yaml.Node.Line says 1 (for "name:"), then reported line is 2. Correct.

	if issues[0].Line != 2 {
		t.Errorf("expected line 2, got %d", issues[0].Line)
	}
}

func TestStyleCheck(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skr-test-style")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	skillPath := filepath.Join(tempDir, skill.SkillFileName)

	// Test bad style
	// 1: ---
	// 2: name: bad-style
	// 3: description: lowercase start.
	// 4: ---
	badStyleContent := `---
name: bad-style
description: lowercase start.
---
`
	if err := os.WriteFile(skillPath, []byte(badStyleContent), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	check := StyleCheck{}
	issues, err := check.Check(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundLowercase := false
	foundPeriod := false
	for _, issue := range issues {
		if issue.Category == CategoryStyle {
			if issue.Message == "description should start with a capital letter" {
				foundLowercase = true
				if issue.Line != 3 {
					t.Errorf("expected line 3 for lowercase warning, got %d", issue.Line)
				}
			}
			if issue.Message == "description should not end with a period" {
				foundPeriod = true
				if issue.Line != 3 {
					t.Errorf("expected line 3 for period warning, got %d", issue.Line)
				}
			}
		}
	}

	if !foundLowercase {
		t.Error("expected lowercase warning")
	}
	if !foundPeriod {
		t.Error("expected period warning")
	}
}
