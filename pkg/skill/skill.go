package skill

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Skill represents the metadata and structure of an Agent Skill
type Skill struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Dependencies []string `yaml:"dependencies,omitempty"`
	Metadata     struct {
		Author  string `yaml:"author,omitempty"`
		Version string `yaml:"version,omitempty"`
	} `yaml:"metadata,omitempty"`
	Path string `yaml:"-"` // Local path to the skill directory
}

const (
	SkillFileName = "SKILL.md"
)

var (
	validNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
)

// Load reads and validates a skill from the given directory path.
func Load(dir string) (*Skill, error) {
	s, err := LoadUnverified(dir)
	if err != nil {
		return nil, err
	}

	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid skill: %w", err)
	}

	return s, nil
}

// LoadUnverified reads a skill from the given directory path without validating it.
// This is useful for installing skills that might have legacy or non-compliant metadata but are otherwise functional.
func LoadUnverified(dir string) (*Skill, error) {
	skillPath := filepath.Join(dir, SkillFileName)

	info, err := os.Stat(skillPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("skill directory must contain a %s file", SkillFileName)
	}
	if err != nil {
		return nil, fmt.Errorf("error checking skill file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s must be a file, not a directory", SkillFileName)
	}

	content, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", SkillFileName, err)
	}

	skill, err := parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s frontmatter: %w", SkillFileName, err)
	}
	skill.Path = dir

	return skill, nil
}

// Validate checks if the skill metadata is valid according to the specification.
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(s.Name) > 64 {
		return fmt.Errorf("name must be 64 characters or less")
	}
	if !validNameRegex.MatchString(s.Name) {
		return fmt.Errorf("name must contain only lowercase alphanumeric characters and hyphens")
	}

	if s.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(s.Description) > 1024 {
		return fmt.Errorf("description must be 1024 characters or less")
	}

	// Validate directory structure matches name (warning or error?)
	// Strictly speaking, the spec says name "Should match the directory name".
	// We won't enforce it as a hard error here but it's good practice.

	return nil
}

func parseFrontmatter(content []byte) (*Skill, error) {
	// Frontmatter is anticipated to be between the first two "---" lines
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, fmt.Errorf("missing frontmatter start delimiter '---'")
	}

	end := bytes.Index(content[4:], []byte("\n---"))
	if end == -1 {
		return nil, fmt.Errorf("missing frontmatter end delimiter '---'")
	}

	// Adjust index to account for skipping the first 4 bytes
	frontmatter := content[4 : 4+end]

	var s Skill
	if err := yaml.Unmarshal(frontmatter, &s); err != nil {
		return nil, err
	}

	return &s, nil
}
