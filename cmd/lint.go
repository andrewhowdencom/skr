package cmd

import (
	"fmt"
	"os"

	"github.com/andrewhowdencom/skr/pkg/lint"
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

		linter := lint.NewLinter([]lint.Check{
			&lint.SpecCheck{},
			&lint.StyleCheck{},
		})

		issues, err := linter.Run(path)
		if err != nil {
			return err
		}

		var formattedOutput string
		switch format {
		case "sarif":
			formattedOutput, err = lint.FormatSARIF(issues)
		case "checkstyle":
			formattedOutput, err = lint.FormatCheckstyle(issues)
		case "gnu":
			formattedOutput, err = lint.FormatGNU(issues)
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
		for _, issue := range issues {
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

		if shouldExit {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	lintCmd.Flags().String("format", "gnu", "Output format (gnu, sarif, checkstyle)")
	lintCmd.Flags().String("output", "", "Output file path (default stdout)")
	lintCmd.Flags().StringSlice("fail-on", []string{"spec", "style"}, "Categories to fail on (spec, style)")
	rootCmd.AddCommand(lintCmd)
}
