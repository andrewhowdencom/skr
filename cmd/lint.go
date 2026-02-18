package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewhowdencom/skr/pkg/lint"
	"github.com/andrewhowdencom/skr/pkg/lint/rules"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint [path]",
	Short: "Lint an Agent Skill",
	Long: `Lint an Agent Skill against specification and style guidelines.

Outputs issues in various formats (GNU, SARIF, Checkstyle).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		format, _ := cmd.Flags().GetString("format")
		outputFile, _ := cmd.Flags().GetString("output")
		failOn, _ := cmd.Flags().GetStringSlice("fail-on")
		fix, _ := cmd.Flags().GetBool("fix")

		linter := lint.NewLinter([]lint.Checker{
			&rules.Name{},
			&rules.Description{},
			&rules.DescriptionCapitalization{},
			&rules.DescriptionPeriod{},
		})

		var allIssues []lint.Issue
		var validationErrors []error

		// Recursive scan
		err := filepath.WalkDir(path, func(currentPath string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}

			// Skip hidden directories and node_modules
			if d.Name() != "." && (d.Name()[0] == '.' || d.Name() == "node_modules") {
				return filepath.SkipDir
			}

			// Check for SKILL.md
			skillPath := filepath.Join(currentPath, "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				// Found a skill, lint it
				issues, err := linter.Run(currentPath, fix)
				if err != nil {
					validationErrors = append(validationErrors, fmt.Errorf("failed to lint %s: %w", currentPath, err))
					return nil // Continue finding others
				}
				allIssues = append(allIssues, issues...)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		if len(validationErrors) > 0 {
			for _, err := range validationErrors {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			// We might still want to output issues if we found some before error?
			// But usually tool errors are separate.
		}

		var formattedOutput string
		switch format {
		case "sarif":
			formattedOutput, err = lint.FormatSARIF(allIssues)
		case "checkstyle":
			formattedOutput, err = lint.FormatCheckstyle(allIssues)
		case "gnu":
			formattedOutput, err = lint.FormatGNU(allIssues)
		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		if err != nil {
			return fmt.Errorf("failed to format issues: %w", err)
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(formattedOutput), 0644); err != nil {
				return fmt.Errorf("failed to write output to file: %w", err)
			}
		} else {
			fmt.Print(formattedOutput)
		}

		// Check for failures based on fail-on categories
		shouldExit := false
		for _, issue := range allIssues {
			if issue.Fixed {
				continue // Don't fail if issue was fixed
			}
			for _, category := range failOn {
				if string(issue.Category) == category {
					shouldExit = true
					break
				}
			}
			if shouldExit {
				break
			}
		}

		if shouldExit || len(validationErrors) > 0 {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	lintCmd.Flags().String("format", "gnu", "Output format (gnu, sarif, checkstyle)")
	lintCmd.Flags().String("output", "", "Output file path (default stdout)")
	lintCmd.Flags().StringSlice("fail-on", []string{"spec", "style"}, "Categories to fail on (spec, style)")
	lintCmd.Flags().Bool("fix", false, "Automatically fix issues where possible")
	rootCmd.AddCommand(lintCmd)
}
