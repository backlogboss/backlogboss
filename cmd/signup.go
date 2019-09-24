package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

// signupCmd represents the signup command
var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Sign up at https://backlogboss.xyz",
	Long: `Signing up and logging in allows you to view projects and tasks on https://backlogboss.xyz/dashboard.
	You also receive daily email nudges to finish current open tasks.
	
	Another advantage is the ability to work in projects across multiple computers.
	
	It's completely optional. Backlogboss continues working exactly same in the logged mode.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Signup()
	},
}

func init() {
	rootCmd.AddCommand(signupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// signupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// signupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
