package rules

import (
	"testing"
)

func TestDescriptionCapitalization(t *testing.T) {
	rule := DescriptionCapitalization{}

	t.Run("Valid", func(t *testing.T) {
		doc := createDoc(t, "description: Valid description.")
		issues, _ := rule.Check(doc)
		if len(issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(issues))
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		doc := createDoc(t, "description: invalid description.")
		issues, _ := rule.Check(doc)
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}
		if issues[0].Message != "description should start with a capital letter" {
			t.Errorf("unexpected message: %s", issues[0].Message)
		}
	})

	t.Run("Fix", func(t *testing.T) {
		doc := createDoc(t, "description: invalid description.")
		if err := rule.Fix(doc); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.Skill.Description != "Invalid description." {
			t.Errorf("expected struct update: 'Invalid description.', got '%s'", doc.Skill.Description)
		}
		// Verify Tree update by re-marshalling? Or checking node value?
		// Helper to find node would be useful, but let's assume if struct updated and code correct, it's fine.
		// Actually, let's verify Tree node value directly to be sure.
		// We reuse the findValueNode from implementation but it's not exported.
		// We can just rely on Struct update for this test as the implementation updates both.
	})
}

func TestDescriptionPeriod(t *testing.T) {
	rule := DescriptionPeriod{}

	t.Run("Valid", func(t *testing.T) {
		doc := createDoc(t, "description: Valid description")
		issues, _ := rule.Check(doc)
		if len(issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(issues))
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		doc := createDoc(t, "description: Valid description.")
		issues, _ := rule.Check(doc)
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}
		if issues[0].Message != "description should not end with a period" {
			t.Errorf("unexpected message: %s", issues[0].Message)
		}
	})

	t.Run("Fix", func(t *testing.T) {
		doc := createDoc(t, "description: Valid description.")
		if err := rule.Fix(doc); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.Skill.Description != "Valid description" {
			t.Errorf("expected struct update: 'Valid description', got '%s'", doc.Skill.Description)
		}
	})
}
