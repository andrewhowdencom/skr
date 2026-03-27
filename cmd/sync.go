package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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

		cfg, err := config.LoadMerged(projectRoot)
		if err != nil {
			return err
		}

		if len(cfg.Skills) == 0 {
			slog.Info("no skills defined in config, nothing to sync")
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

		// 4. Load Lock file
		var lockFile *config.LockFile
		lockFilePath, err := config.FindLockFile(projectRoot)
		if err == nil {
			lockFile, _ = config.LoadLock(lockFilePath)
		} else {
			lockFile = config.NewLockFile()
			lockFilePath = filepath.Join(projectRoot, config.AltLockFileName)
		}

		// 5. Install missing skills
		for _, ref := range cfg.Skills {
			installRef := ref

			// If we have a locked digest and it's not already embedded in the ref
			if lockedDigest, ok := lockFile.Skills[ref]; ok && !strings.Contains(ref, "@") && lockedDigest != "" {
				// Strip tag if it exists when adding digest, because oras expects either tag OR digest.
				// Ref format: domain/repo:tag
				// If we append @digest to domain/repo:tag, containerd/docker accepts it (usually resolves to digest).
				// We can simply pass the digest replacing the tag if we wanted to, or append it.
				// The image spec allows namespace/name:tag@digest.
				installRef = ref + "@" + lockedDigest
			}

			slog.Info("syncing skill", "ref", ref, "installRef", installRef)

			// Install using the action package
			_, digest, err := action.InstallSkill(ctx, st, installRef, installRoot, false)
			if err != nil {
				return fmt.Errorf("failed to install %s: %w", ref, err)
			}

			lockFile.Skills[ref] = digest
		}

		// 6. Prune and Save Lock file
		validSkills := make(map[string]bool)
		for _, ref := range cfg.Skills {
			validSkills[ref] = true
		}
		for lockedRef := range lockFile.Skills {
			if !validSkills[lockedRef] {
				delete(lockFile.Skills, lockedRef)
			}
		}

		if err := lockFile.SaveTo(lockFilePath); err != nil {
			return fmt.Errorf("failed to save lock file: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
