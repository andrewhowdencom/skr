package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/config"
	"github.com/spf13/cobra"
)

var (
	initAgent string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new skr project",
	Long: `Initialize a new skr project in the current directory.

This command creates a .skr.yaml configuration file and the necessary
directory structure for agent skills (e.g., .agent/skills).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Check if config already exists
		existingConfig, err := config.FindConfigFile(cwd)
		if err == nil {
			return fmt.Errorf("configuration file already exists at: %s", existingConfig)
		}

		// Create .agent/skills
		agentSkillsDir := filepath.Join(cwd, ".agent", "skills")
		if err := os.MkdirAll(agentSkillsDir, 0755); err != nil {
			return fmt.Errorf("failed to create skills directory: %w", err)
		}

		// Create .skr.yaml
		cfg := &config.Config{
			Agents: []string{initAgent},
			Skills: []string{},
		}

		// Save uses SaveTo internally with .skr.yaml
		// We can just use Save(cwd)
		if err := cfg.Save(cwd); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Initialized skr project in %s\n", cwd)
		fmt.Printf("Created %s\n", agentSkillsDir)
		fmt.Printf("Created .skr.yaml with agent: %s\n", initAgent)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initAgent, "agent", "a", "antigravity", "Agent to configure for (e.g. antigravity, standard)")
}
