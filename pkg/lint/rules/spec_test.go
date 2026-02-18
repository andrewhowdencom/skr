package rules

import (
	"testing"

	"github.com/andrewhowdencom/skr/pkg/lint"
	"github.com/andrewhowdencom/skr/pkg/skill"
	"gopkg.in/yaml.v3"
)

func createDoc(t *testing.T, content string) *lint.Document {
	doc := &lint.Document{
		Path:    "test/SKILL.md",
		Content: []byte(content),
	}
	// Manual parse setup for testing to avoid file I/O
	var root yaml.Node
	if err := yaml.Unmarshal(doc.Content, &root); err != nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}
	doc.Tree = &root

	var s skill.Skill
	if err := yaml.Unmarshal(doc.Content, &s); err != nil {
		t.Fatalf("failed to Unmarshal skll: %v", err)
	}
	doc.Skill = &s
	return doc
}

func TestNameRule(t *testing.T) {
	rule := Name{}

	t.Run("Valid Name", func(t *testing.T) {
		doc := createDoc(t, "name: valid-name")
		issues, _ := rule.Check(doc)
		if len(issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(issues))
		}
	})

	t.Run("Invalid Name Caps", func(t *testing.T) {
		doc := createDoc(t, "name: InvalidName")
		issues, _ := rule.Check(doc)
		if len(issues) != 1 {
			t.Errorf("expected 1 issue, got %d", len(issues))
		} else {
			if issues[0].Message != "name must contain only lowercase alphanumeric characters and hyphens" {
				t.Errorf("unexpected message: %s", issues[0].Message)
			}
		}
	})

	t.Run("Missing Name", func(t *testing.T) {
		doc := createDoc(t, "description: foo")
		issues, _ := rule.Check(doc)
		if len(issues) != 1 {
			t.Errorf("expected 1 issue, got %d", len(issues))
		} else {
			if issues[0].Message != "name is required" {
				t.Errorf("unexpected message: %s", issues[0].Message)
			}
		}
	})
}
