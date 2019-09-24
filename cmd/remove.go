package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes all local projects settings and data",
	Long:  `Removes all local projects settings and data`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.DeleteData()
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
