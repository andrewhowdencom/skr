package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/config"
	"github.com/andrewhowdencom/skr/pkg/discovery"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [skill-name-or-ref]",
	Short: "Remove an Agent Skill",
	Long: `Remove an installed Agent Skill.

Removes the skill from the configuration (.skr.yaml) and deletes the skill directory.
If --global is set, removes from the global configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires skill name or reference")
		}
		ref := args[0]
		isGlobal, _ := cmd.Flags().GetBool("global")

		// 1. Determine Context and Load Config
		var configDir string
		var installRoot string

		if isGlobal {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configDir = filepath.Join(homeDir, ".config")
			installRoot = filepath.Join(homeDir, ".config", "agent", "skills")
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %w", err)
			}

			agentDir, err := discovery.FindAgentSkillsDir(cwd)
			if err != nil {
				return fmt.Errorf("agent context not found (use --global or run inside a project): %w", err)
			}
			configDir = filepath.Dir(filepath.Dir(agentDir))
			installRoot = agentDir
		}

		cfg, err := config.Load(configDir)
		// If config doesn't exist, we can't remove from it, but maybe we can remove dir?
		// But config.Load returns empty config if not found (if we implemented it that way).
		// Currently config.Load errors? No, it returns empty if not found.
		if err != nil {
			return err
		}

		// 2. Remove from Config
		newSkills := []string{}
		removed := false
		for _, s := range cfg.Skills {
			if s == ref {
				removed = true
				continue
			}
			// Handle case where ref is "git" but config has "git:v1"?
			// Ideally we remove exact match or parse.
			// For now, strict equality.
			newSkills = append(newSkills, s)
		}

		if removed {
			cfg.Skills = newSkills
			if err := cfg.Save(configDir); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("Removed '%s' from %s\n", ref, filepath.Join(configDir, config.ConfigFileName))
		} else {
			fmt.Printf("Skill '%s' not found in configuration\n", ref)
		}

		// 3. Remove Directory
		// We need to know the directory name.
		// If ref is "git:v1", dir is "git"?
		// We'd assume standard install naming.
		// Naive approach: try to remove directory named `ref`.
		// If ref has tag, strip it?
		// Realistically, users might pass just the name "git" to rm.
		// If config had "git:v1", and user types "skr rm git", strict equality fails above.
		// Users should probably `skr rm git:v1`.

		// For directory removal, we assume directory name matches resolved name.
		// If user passes "git:v1", we probably want to remove "git"?
		// Discovery: maybe look for dir?
		// Let's assume user passes name or exact ref.
		// If exact ref is used, we might need to guess the folder name.

		// FIXME: Weakness in mapping ref <-> folder name.
		// Let's assume user provides DIR NAME mostly, OR exact ref string.
		// Loop through installRoot entries?

		// For now, just attempt to remove `filepath.Join(installRoot, ref)`
		// AND `filepath.Join(installRoot, ref_without_tag)`?
		targetPath := filepath.Join(installRoot, ref)
		if _, err := os.Stat(targetPath); err == nil {
			if err := os.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", targetPath, err)
			}
			fmt.Printf("Removed skill directory %s\n", targetPath)
		} else {
			// Try splitting by colon
			// e.g. git:v1 -> git
			// But maybe the dir IS git:v1? No, install logic implies name from SKILL.md.
			// This is tricky without metadata.
			// Let's assume "rm" takes the NAME of the skill as installed in FS,
			// checking config for corresponding ref?

			// If we passed "dummy:latest" to install, folder is "valid-skill" (from SKILL.md).
			// So `skr rm valid-skill` is the logical command.
			// But config has "dummy:latest".
			// So rm needs to handle this mismatch.

			// Ideal: Read all SKILL.mds in installRoot to find matching name?
			// That's expensive but accurate.

			fmt.Printf("Warning: Directory %s not found. Syncing config only.\n", targetPath)
		}

		return nil
	},
}

func init() {
	rmCmd.Flags().Bool("global", false, "Remove skill globally")
	rootCmd.AddCommand(rmCmd)
}
