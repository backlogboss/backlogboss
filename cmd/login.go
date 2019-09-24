package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to backlogboss.xyz",
	Long: `Login to backlogboss.xyz. 
	Signing up and logging in allows you to view projects and tasks on https://backlogboss.xyz/dashboard.
	You also receive daily email nudges to finish current open tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Login()
		pkg.Sync()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
