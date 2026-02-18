package cmd

import (
	"fmt"

	"github.com/andrewhowdencom/skr/pkg/lint"
	"github.com/andrewhowdencom/skr/pkg/skill"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate an Agent Skill definition",
	Long: `Validate an Agent Skill's integrity and adherence to the specification.

Checks for:
- Existence of SKILL.md
- Valid frontmatter
- Spec compliance (naming, fields)
- Directory structure

If [path] is not provided, defaults to the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		linter := lint.NewLinter([]lint.Check{
			&lint.SpecCheck{},
		})

		issues, err := linter.Run(path)
		if err != nil {
			return fmt.Errorf("skill is invalid: %w", err)
		}

		hasError := false
		for _, issue := range issues {
			if issue.Category == lint.CategorySpec {
				fmt.Printf("Error: %s\n", issue.Message)
				hasError = true
			}
		}

		if hasError {
			return fmt.Errorf("skill is invalid")
		}

		// We need to load the skill to get the name for the success message
		// efficiently, or just say "Skill is valid" without the name.
		// Let's load it just for the name if valid.
		s, err := skill.LoadUnverified(path)
		if err == nil {
			fmt.Printf("Skill '%s' is valid.\n", s.Name)
		} else {
			fmt.Println("Skill is valid.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
