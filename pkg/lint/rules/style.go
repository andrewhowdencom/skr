package rules

import (
	"unicode"

	"github.com/andrewhowdencom/skr/pkg/lint"
)

type DescriptionCapitalization struct{}

func (r *DescriptionCapitalization) Name() string {
	return "style/description-capitalization"
}

func (r *DescriptionCapitalization) Check(doc *lint.Document) ([]lint.Issue, error) {
	var issues []lint.Issue
	if doc.Skill == nil || doc.Skill.Description == "" {
		return nil, nil
	}

	desc := doc.Skill.Description
	firstChar := rune(desc[0])
	if unicode.IsLower(firstChar) {
		valNode := findValueNode(doc.Tree, "description")
		line, col := 0, 0
		if valNode != nil {
			line = valNode.Line
			col = valNode.Column
		}
		issues = append(issues, lint.Issue{
			Path:     doc.Path,
			Line:     line,
			Column:   col,
			Message:  "description should start with a capital letter",
			Category: lint.CategoryStyle,
		})
	}
	return issues, nil
}

func (r *DescriptionCapitalization) Fix(doc *lint.Document) error {
	if doc.Skill == nil || doc.Skill.Description == "" {
		return nil
	}
	desc := doc.Skill.Description
	runes := []rune(desc)
	if len(runes) > 0 && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
		newDesc := string(runes)

		// Update struct
		doc.Skill.Description = newDesc

		// Update Tree
		valNode := findValueNode(doc.Tree, "description")
		if valNode != nil {
			valNode.Value = newDesc
			doc.MarkDirty()
		}
	}
	return nil
}

type DescriptionPeriod struct{}

func (r *DescriptionPeriod) Name() string {
	return "style/description-period"
}

func (r *DescriptionPeriod) Check(doc *lint.Document) ([]lint.Issue, error) {
	var issues []lint.Issue
	if doc.Skill == nil || doc.Skill.Description == "" {
		return nil, nil
	}

	desc := doc.Skill.Description
	if desc[len(desc)-1] == '.' {
		valNode := findValueNode(doc.Tree, "description")
		line, col := 0, 0
		if valNode != nil {
			line = valNode.Line
			col = valNode.Column
		}
		issues = append(issues, lint.Issue{
			Path:     doc.Path,
			Line:     line,
			Column:   col,
			Message:  "description should not end with a period",
			Category: lint.CategoryStyle,
		})
	}
	return issues, nil
}

func (r *DescriptionPeriod) Fix(doc *lint.Document) error {
	if doc.Skill == nil || doc.Skill.Description == "" {
		return nil
	}
	desc := doc.Skill.Description
	if desc[len(desc)-1] == '.' {
		newDesc := desc[:len(desc)-1]

		// Update struct
		doc.Skill.Description = newDesc

		// Update Tree
		valNode := findValueNode(doc.Tree, "description")
		if valNode != nil {
			valNode.Value = newDesc
			doc.MarkDirty()
		}
	}
	return nil
}
