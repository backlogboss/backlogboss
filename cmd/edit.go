package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit current task",
	Long:  `This task opens a text $EDITOR to allow editing the task log `,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Sync()
		messageFlag := cmd.Flag("message")
		if messageFlag.Changed {
			pkg.EditAppend(messageFlag.Value.String())
		} else if cmd.Flag("title").Changed {
			pkg.EditTitle(cmd.Flag("title").Value.String())
		} else {
			pkg.EditTask()
		}

		pkg.Sync()
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringP("message", "m", "", "Append message to the task")
	editCmd.Flags().StringP("title", "t", "", "Update task title(between 10 and 100 characters")
}
