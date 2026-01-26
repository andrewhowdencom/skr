package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/andrewhowdencom/skr/pkg/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login [server]",
	Short: "Log in to an OCI registry",
	Long: `Log in to an OCI registry using credentials.

This establishes an authenticated session for pushing and pulling private skills.
Credentials are stored locally in the user's config directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := "ghcr.io" // Default
		if len(args) > 0 {
			server = args[0]
		}

		// Read flags
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		passwordStdin, _ := cmd.Flags().GetBool("password-stdin")

		// Interactive or Stdin prompt
		if passwordStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			password = strings.TrimSuffix(string(data), "\n")
			password = strings.TrimSuffix(password, "\r")
		} else {
			if username == "" {
				fmt.Print("Username: ")
				fmt.Scanln(&username)
			}
			if password == "" {
				fmt.Print("Password: ")
				bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return err
				}
				fmt.Println() // Newline
				password = string(bytePassword)
			}
		}

		if username == "" || password == "" {
			return fmt.Errorf("username and password required")
		}

		fmt.Printf("Logging into %s as %s...\n", server, username)
		if err := auth.Login(server, username, password); err != nil {
			return err
		}

		fmt.Println("Login Succeeded")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username")
	loginCmd.Flags().StringP("password", "p", "", "Password")
	loginCmd.Flags().Bool("password-stdin", false, "Read password from stdin")
	registryCmd.AddCommand(loginCmd)
}
