package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/skill"
	"gopkg.in/yaml.v3"
)

// Document represents the state of a skill file being linted.
type Document struct {
	Path        string
	Content     []byte
	Tree        *yaml.Node
	Skill       *skill.Skill
	isDirty     bool
	frontmatter []byte // Cache frontmatter for rewriting if needed
	fmStart     int
	fmEnd       int
}

// NewDocument loads a skill file and parses it.
func NewDocument(path string) (*Document, error) {
	skillPath := filepath.Join(path, skill.SkillFileName)
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", skill.SkillFileName, err)
	}

	doc := &Document{
		Path:    skillPath,
		Content: content,
	}

	if err := doc.parse(); err != nil {
		return nil, err
	}

	return doc, nil
}

func (d *Document) parse() error {
	// Simple frontmatter extraction (same as in checks.go)
	if !bytes.HasPrefix(d.Content, []byte("---\n")) {
		return fmt.Errorf("missing frontmatter start delimiter")
	}
	end := bytes.Index(d.Content[4:], []byte("\n---"))
	if end == -1 {
		return fmt.Errorf("missing frontmatter end delimiter")
	}

	d.fmStart = 0
	// The frontmatter block ends after the closing delimiter "\n---".
	// The content[4:] slice starts after the initial "---\n".
	// 'end' is the index of "\n---" within content[4:].
	// So the absolute index of "\n---" is 4 + end.
	// The delimiter itself has length 4.
	// Thus, the frontmatter block ends at: 4 + end + 4.
	d.fmEnd = 4 + end + 4

	d.frontmatter = d.Content[4 : 4+end]

	var root yaml.Node
	if err := yaml.Unmarshal(d.frontmatter, &root); err != nil {
		return fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}
	d.Tree = &root

	var s skill.Skill
	if err := yaml.Unmarshal(d.frontmatter, &s); err != nil {
		return fmt.Errorf("failed to unmarshal into Skill struct: %w", err)
	}
	d.Skill = &s

	return nil
}

// MarkDirty marks the document as modified.
func (d *Document) MarkDirty() {
	d.isDirty = true
}

// Flush writes the document back to disk if it is dirty.
func (d *Document) Flush() error {
	if !d.isDirty {
		return nil
	}

	// Re-serialize YAML tree
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	// d.Tree is a DocumentNode. We encode its first Content item (the MappingNode)
	// to get the raw YAML values without the document separators.
	if err := enc.Encode(d.Tree.Content[0]); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	newFM := buf.Bytes()
	// Encode adds a newline at the end.

	// Construct new content
	// We need to preserve "---\n" at start and "\n---" at end.
	// But `newFM` might effectively be the content between them.

	// Reconstruct the file content:
	// 1. Initial delimiter
	// 2. New Frontmatter
	// 3. Closing delimiter
	// 4. Original content (if any)
	var newContent bytes.Buffer
	newContent.WriteString("---\n")
	newContent.Write(newFM)
	newContent.WriteString("---\n")

	// Wait, original file might have content after frontmatter?
	// SKILL.md is usually just metadata, but spec allows content?
	// The current implementation assumes checks only look at SKILL.md.
	// If SKILL.md has body content, we should preserve it.

	rest := d.Content[d.fmEnd:]
	// Actually d.fmEnd calculation above:
	// content[4:]
	// end index of "\n---"
	// so content[4+end] is start of "\n---"
	// we want to append rest after "\n---" + 4?
	// or does our replacement include the delimiters?

	// Let's assume we replace the whole block from 0 to 4+end+4 with new frontmatter block.

	// NOTE: yaml.v3 Encoder tends to be opinionated. Ideally we'd edit the standard `d.Content` byte slice directly if we knew offsets.
	// But `yaml.Node` doesn't support easy writing back with preservation of comments/style perfectly unless we use `yaml.Encoder`.
	// For now, re-generating the frontmatter is acceptable for "fixing".

	if len(rest) > 0 {
		// If rest starts with newline (it should if we cut after ---), skip it if we added one?
		// \n--- ends the block.
		// If we stripped \n--- from frontmatter slice, fine.
		// We are rebuilding: ---\n + newYAML + ---\n + rest.
		// If rest started immediately after \n---, it might be empty or start with newline.
		// Let's look at `d.fmEnd`:
		// content[4:] -> "name: foo\n..."
		// end -> index of start of "\n---" in that slice.
		// So content[4+end] is '\n'.
		// The delimiter is "\n---". Length 4.
		// So block ends at 4 + end + 4.
		// d.Content[4+end+4:] is the rest.

		// If original was:
		// ---\n
		// foo: bar\n
		// ---\n
		// # content

		// fmEnd points to after last \n.
		newContent.Write(d.Content[d.fmEnd:])
	}

	if err := os.WriteFile(d.Path, newContent.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	d.isDirty = false
	// Re-parse to update offsets and structs?
	d.Content = newContent.Bytes()
	return d.parse()
}
