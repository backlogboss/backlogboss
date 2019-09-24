package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Opens the web dashboard",
	Long:  `Opens the web dashboard`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Web()
	},
}

func init() {
	rootCmd.AddCommand(webCmd)
}
