package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

var commitSwap bool

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Close the current open task and create a new one.",
	Long:  `Close the current open task and create a new one.'`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Sync()
		pkg.CommitTask()
		pkg.Sync()
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
