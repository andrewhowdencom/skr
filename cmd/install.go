package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/action"
	"github.com/andrewhowdencom/skr/pkg/config"
	"github.com/andrewhowdencom/skr/pkg/discovery"
	"github.com/andrewhowdencom/skr/pkg/store"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [ref]",
	Short: "Install an Agent Skill",
	Long: `Install an Agent Skill.

Adds the skill to the configuration (.skr.yaml) and synchronizes the installation.
If --global is set, installs to the global configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires skill reference (e.g. tag or digest)")
		}
		ref := args[0]
		isGlobal, _ := cmd.Flags().GetBool("global")
		ctx := cmd.Context()

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

			// Find Project Root
			agentDir, err := discovery.FindAgentSkillsDir(cwd)
			if err != nil {
				// If not found, default to cwd as root? Or fail?
				// For init, we might want cwd. For install, we usually want existing.
				// Let's assume cwd if fail, but warn?
				// Actually, discovery fails if .agent/skills doesn't exist.
				// So we should strictly fail unless --global, OR rely on `skr init` to create struct.
				// But user might run install in empty dir.
				// Let's error for now.
				return fmt.Errorf("agent context not found (use --global or run inside a project): %w", err)
			}
			configDir = filepath.Dir(filepath.Dir(agentDir)) // Parent of .agent
			installRoot = agentDir
		}

		// Ensure install root exists
		if err := os.MkdirAll(installRoot, 0755); err != nil {
			return fmt.Errorf("failed to create install directory: %w", err)
		}
		// Ensure config dir exists (mostly for global)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		cfg, err := config.Load(configDir)
		if err != nil {
			return err
		}

		// 2. Add to Config
		// Check if already exists?
		exists := false
		for _, s := range cfg.Skills {
			if s == ref {
				exists = true
				break
			}
		}
		if !exists {
			cfg.Skills = append(cfg.Skills, ref)
			if err := cfg.Save(configDir); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("Added '%s' to %s\n", ref, filepath.Join(configDir, config.ConfigFileName))
		} else {
			fmt.Printf("Skill '%s' already in config\n", ref)
		}

		// 3. Perform Install (Sync)
		// We could call sync command, or just install this one skill.
		// For efficiency, let's just install this one.
		// But strictly "Sync" implies ensuring everything.
		// Let's just install this one for now to be fast.

		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		fmt.Printf("Installing %s to %s...\n", ref, installRoot)
		name, err := action.InstallSkill(ctx, st, ref, installRoot)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully installed '%s' (%s)\n", name, ref)
		return nil
	},
}

func init() {
	installCmd.Flags().Bool("global", false, "Install skill globally")
	rootCmd.AddCommand(installCmd)
}
