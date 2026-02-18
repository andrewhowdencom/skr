package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"gopkg.in/yaml.v3"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
)

type SpecCheck struct{}

func (c *SpecCheck) Check(dir string) ([]Issue, error) {
	var issues []Issue
	skillPath := filepath.Join(dir, skill.SkillFileName)

	info, err := os.Stat(skillPath)
	if os.IsNotExist(err) {
		issues = append(issues, Issue{
			Path:     dir,
			Message:  fmt.Sprintf("skill directory must contain a %s file", skill.SkillFileName),
			Category: CategorySpec,
		})
		return issues, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error checking skill file: %w", err)
	}
	if info.IsDir() {
		issues = append(issues, Issue{
			Path:     skillPath,
			Message:  fmt.Sprintf("%s must be a file, not a directory", skill.SkillFileName),
			Category: CategorySpec,
		})
		return issues, nil
	}

	content, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", skill.SkillFileName, err)
	}

	// Validate Frontmatter existence
	if !bytes.HasPrefix(content, []byte("---\n")) {
		issues = append(issues, Issue{
			Path:     skillPath,
			Line:     1,
			Message:  "missing frontmatter start delimiter '---'",
			Category: CategorySpec,
		})
		return issues, nil
	}

	end := bytes.Index(content[4:], []byte("\n---"))
	if end == -1 {
		issues = append(issues, Issue{
			Path:     skillPath,
			Message:  "missing frontmatter end delimiter '---'",
			Category: CategorySpec,
		})
		return issues, nil
	}

	frontmatter := content[4 : 4+end]

	// Parse into yaml.Node to get line numbers
	var root yaml.Node
	if err := yaml.Unmarshal(frontmatter, &root); err != nil {
		issues = append(issues, Issue{
			Path:     skillPath,
			Message:  fmt.Sprintf("failed to parse YAML frontmatter: %v", err),
			Category: CategorySpec,
		})
		return issues, nil
	}

	// Helper to find a node by key in the top-level mapping
	findNode := func(key string) *yaml.Node {
		if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
			mapping := root.Content[0]
			if mapping.Kind == yaml.MappingNode {
				for i := 0; i < len(mapping.Content); i += 2 {
					if mapping.Content[i].Value == key {
						// Return the value node, but we might want the key node for some errors?
						// Usually validating the value, so return value node.
						// But if valid, we return value node.
						return mapping.Content[i+1]
					}
				}
			}
		}
		return nil
	}

	// Also parse into struct for easier logic checks (optional, but keep it simple)
	var s skill.Skill
	if err := yaml.Unmarshal(frontmatter, &s); err != nil {
		// Should have been caught above, but just in case
		return issues, nil
	}

	// Validate Name
	nameNode := findNode("name")
	if s.Name == "" {
		line := 1
		if nameNode != nil {
			line = nameNode.Line
		} else {
			// If missing, arguably point to parsing start or file start.
			// Since we unmarshal frontmatter as a snippet, lines are relative to start of snippet + offset.
			// Frontmatter starts at line 1 (the first ---) + 1.
			// yaml.Unmarshal on the slice 'frontmatter' will report lines relative to line 1 of that slice.
			// So we need to add offset: 4 bytes skipped means we are effectively at line 2 of file.
			// Actually yaml.Node.Line is 1-based.
			// Let's adjust: The slice starts at file line 2.
		}
		issues = append(issues, Issue{
			Path:     skillPath,
			Line:     line + 1, // Offset for frontmatter start
			Message:  "name is required",
			Category: CategorySpec,
		})
	} else {
		line := nameNode.Line + 1
		col := nameNode.Column
		if len(s.Name) > 64 {
			issues = append(issues, Issue{
				Path:     skillPath,
				Line:     line,
				Column:   col,
				Message:  "name must be 64 characters or less",
				Category: CategorySpec,
			})
		}
		if !validNameRegex.MatchString(s.Name) {
			issues = append(issues, Issue{
				Path:     skillPath,
				Line:     line,
				Column:   col,
				Message:  "name must contain only lowercase alphanumeric characters and hyphens",
				Category: CategorySpec,
			})
		}
	}

	// Validate Description
	descNode := findNode("description")
	if s.Description == "" {
		line := 1
		if descNode != nil {
			line = descNode.Line
		}
		issues = append(issues, Issue{
			Path:     skillPath,
			Line:     line + 1,
			Message:  "description is required",
			Category: CategorySpec,
		})
	} else {
		line := descNode.Line + 1
		col := descNode.Column
		if len(s.Description) > 1024 {
			issues = append(issues, Issue{
				Path:     skillPath,
				Line:     line,
				Column:   col,
				Message:  "description must be 1024 characters or less",
				Category: CategorySpec,
			})
		}
	}

	return issues, nil
}

type StyleCheck struct{}

func (c *StyleCheck) Check(dir string) ([]Issue, error) {
	var issues []Issue
	skillPath := filepath.Join(dir, skill.SkillFileName)

	content, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, nil
	}

	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, nil
	}
	end := bytes.Index(content[4:], []byte("\n---"))
	if end == -1 {
		return nil, nil
	}
	frontmatter := content[4 : 4+end]

	var root yaml.Node
	if err := yaml.Unmarshal(frontmatter, &root); err != nil {
		return nil, nil
	}

	var s skill.Skill
	if err := yaml.Unmarshal(frontmatter, &s); err != nil {
		return nil, nil
	}

	findNode := func(key string) *yaml.Node {
		if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
			mapping := root.Content[0]
			if mapping.Kind == yaml.MappingNode {
				for i := 0; i < len(mapping.Content); i += 2 {
					if mapping.Content[i].Value == key {
						return mapping.Content[i+1]
					}
				}
			}
		}
		return nil
	}

	descNode := findNode("description")
	if descNode != nil && len(s.Description) > 0 {
		line := descNode.Line + 1 // Offset
		col := descNode.Column

		firstChar := rune(s.Description[0])
		if firstChar >= 'a' && firstChar <= 'z' {
			issues = append(issues, Issue{
				Path:     skillPath,
				Line:     line,
				Column:   col,
				Message:  "description should start with a capital letter",
				Category: CategoryStyle,
			})
		}
		if s.Description[len(s.Description)-1] == '.' {
			issues = append(issues, Issue{
				Path:     skillPath,
				Line:     line,
				Column:   col,
				Message:  "description should not end with a period",
				Category: CategoryStyle,
			})
		}
	}

	return issues, nil
}
