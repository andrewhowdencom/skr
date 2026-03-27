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

var updateCmd = &cobra.Command{
	Use:   "update [skill]",
	Short: "Update installed skills to their latest versions",
	Long: `Update looks up the skills declared in .skr.yaml and fetches the newest version of their manifest from the remote registry, bypassing the local cache/lock.

If a new digest is discovered (e.g., when a floating tag like :latest points to a new build), .skr.lock is updated automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		updateAll, _ := cmd.Flags().GetBool("all")
		if !updateAll && len(args) == 0 {
			return fmt.Errorf("requires skill reference or --all flag")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}

		projectRoot := cwd
		agentDir, err := discovery.FindAgentSkillsDir(cwd)
		if err == nil {
			projectRoot = filepath.Dir(filepath.Dir(agentDir))
		}

		cfg, err := config.LoadMerged(projectRoot)
		if err != nil {
			return err
		}

		if len(cfg.Skills) == 0 {
			slog.Info("no skills defined in config, nothing to update")
			return nil
		}

		ctx := cmd.Context()
		st, err := store.New("")
		if err != nil {
			return fmt.Errorf("failed to initialize store: %w", err)
		}

		installRoot := filepath.Join(projectRoot, ".agent", "skills")
		if err := os.MkdirAll(installRoot, 0755); err != nil {
			return fmt.Errorf("failed to create install root %s: %w", installRoot, err)
		}

		var lockFile *config.LockFile
		lockFilePath, err := config.FindLockFile(projectRoot)
		if err == nil {
			lockFile, _ = config.LoadLock(lockFilePath)
		} else {
			lockFile = config.NewLockFile()
			lockFilePath = filepath.Join(projectRoot, config.AltLockFileName)
		}

		var skillsToUpdate []string
		if updateAll {
			skillsToUpdate = cfg.Skills
		} else {
			target := args[0]
			for _, ref := range cfg.Skills {
				if stripTag(ref) == target || ref == target {
					skillsToUpdate = append(skillsToUpdate, ref)
				}
			}
			if len(skillsToUpdate) == 0 {
				return fmt.Errorf("skill '%s' not found in .skr.yaml", target)
			}
		}

		for _, ref := range skillsToUpdate {
			slog.Info("updating skill", "ref", ref)

			// We force a pull so we pass the original ref, NOT combining with the lock digest
			_, digest, err := action.InstallSkill(ctx, st, ref, installRoot, true)
			if err != nil {
				return fmt.Errorf("failed to update %s: %w", ref, err)
			}

			lockFile.Skills[ref] = digest
		}

		// Prune lock file against current configuration
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

func stripTag(ref string) string {
	idx := strings.LastIndex(ref, ":")
	if idx != -1 {
		if !strings.Contains(ref[idx:], "/") {
			return ref[:idx]
		}
	}
	return ref
}

func init() {
	updateCmd.Flags().Bool("all", false, "Update all skills listed in .skr.yaml")
	rootCmd.AddCommand(updateCmd)
}
