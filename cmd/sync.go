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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Reconcile installed skills with .skr.yaml",
	Long: `Synchonize the installed skills in .agent/skills with the declarative list in .skr.yaml.

- Installs skills listed in .skr.yaml that are missing from .agent/skills.
- Removes skills in .agent/skills that are not present in .skr.yaml (unless they are local dependencies/ignored, TBD).
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}

		// 1. Load Config
		// We expect .skr.yaml to be in the project root.
		// Discovery logic finds .agent/skills, essentially finding the project root.
		// So let's find project root first.
		projectRoot := cwd
		agentDir, err := discovery.FindAgentSkillsDir(cwd)
		if err == nil {
			// If .agent/skills found, assume project root is parent of .agent
			projectRoot = filepath.Dir(filepath.Dir(agentDir))
		}
		// If not found, try to load config in cwd anyway, essentially treating cwd as root?
		// Or if discovery failed, maybe we are initializing?
		// But sync implies existing structure.

		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		if len(cfg.Skills) == 0 {
			fmt.Println("No skills defined in .skr.yaml. Nothing to sync.")
			// Should strictly remove everything? For safety, maybe just warn for now.
			return nil
		}

		// 2. Initialize Store
		ctx := cmd.Context()
		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		// 3. Ensure .agent/skills exists
		installRoot := filepath.Join(projectRoot, ".agent", "skills")
		if err := os.MkdirAll(installRoot, 0755); err != nil {
			return fmt.Errorf("failed to create install root %s: %w", installRoot, err)
		}

		// 4. Install missing skills
		// Naive implementation: iterate config.Skills, checked if installed.
		// NOTE: config.Skills might be "git:v1" or just "git".
		// We need to resolve name.
		for _, ref := range cfg.Skills {
			// Check if installed
			// We need to parse name from ref?
			// If ref is "git:v1", name is "git".
			// If ref is "ghcr.io/mypkg/git:latest", name is "git"?
			// Ideally config should map name -> ref, OR we parse ref.
			// Simple parsing: last component before tag.
			// For now, let's assume ref is compatible with what we expect.
			// Wait, how do we know if it's already installed?
			// By checking directory existence?

			// FIXME: name resolution is tricky without pulling.
			// Let's rely on 'install' logic which unpacks to temp and gets name.
			// For sync, efficiency matters.
			// IF we mandate ref format <name>:<tag> or just <name>, we can guess.
			// But for now, let's just attempt install if NOT skipping?

			// Actually, let's call the install logic directly.
			// Reusing logic from installCmd would be good.
			// Moving install logic to a pkg/action or similar?
			// For now, inline or copy/paste logic from install.go.

			fmt.Printf("Syncing skill: %s\n", ref)

			// Install using the action package
			_, err := action.InstallSkill(ctx, st, ref, installRoot)
			if err != nil {
				return fmt.Errorf("failed to install %s: %w", ref, err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
