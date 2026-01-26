package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/skill"
)

// FindAgentSkillsDir traverses upwards from startDir to find .agent/skills
func FindAgentSkillsDir(startDir string) (string, error) {
	dir := startDir
	for i := 0; i < 100; i++ { // Limit to 100 levels to prevent infinite loops/too deep
		target := filepath.Join(dir, ".agent", "skills")
		info, err := os.Stat(target)
		if err == nil && info.IsDir() {
			return target, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find .agent/skills directory in any parent of %s", startDir)
}

// InstalledSkill represents a skill found in the agent's environment
type InstalledSkill struct {
	Name     string
	Path     string
	Version  string // From SKILL.md if available, else "unknown" (or we might parse it)
	IsGlobal bool
}

// ListInstalledSkills discovers all skills in the .agent/skills directory accessible from startDir
func ListInstalledSkills(startDir string) ([]InstalledSkill, error) {
	var skills []InstalledSkill

	// 1. Local Agent Skills
	skillsDir, err := FindAgentSkillsDir(startDir)
	if err == nil { // It's okay if not found, just return empty (or global only)
		entries, err := os.ReadDir(skillsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					skillPath := filepath.Join(skillsDir, entry.Name())
					s, err := skill.Load(skillPath)
					if err == nil {
						skills = append(skills, InstalledSkill{
							Name:     s.Name,
							Path:     skillPath,
							Version:  "local", // TODO: Parse version from SKILL.md content if added to spec
							IsGlobal: false,
						})
					}
				}
			}
		}
	}

	// 2. Global Skills
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalDir := filepath.Join(homeDir, ".config", "agent", "skills")
		entries, err := os.ReadDir(globalDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					skillPath := filepath.Join(globalDir, entry.Name())
					s, err := skill.Load(skillPath)
					if err == nil {
						// Check if overridden by local
						overridden := false
						for _, local := range skills {
							if local.Name == s.Name {
								overridden = true
								break
							}
						}

						if !overridden {
							skills = append(skills, InstalledSkill{
								Name:     s.Name,
								Path:     skillPath,
								Version:  "valid", // TODO: Version parsing
								IsGlobal: true,
							})
						}
					}
				}
			}
		}
	}

	return skills, nil
}
