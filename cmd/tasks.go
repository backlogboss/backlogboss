package cmd

import (
	"backlogboss/pkg"

	"github.com/spf13/cobra"
)

// tasksCmd represents the tasks command
var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Displays the list of open and queued tasks. Select a task to edit, swap, commit or delete.",
	Long: `
	Displays the list of tasks. 
	Use "tasks open" to show only open tasks or tasks open,queued to filter tasks.
	
	Filters: open, queued, done, all
	Default: open, queued
	
	Select a task to swap, commit or edit.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		pkg.Sync()
		if cmd.Flag(pkg.All.String()).Changed {
			pkg.Log(pkg.All)
		} else if cmd.Flag(pkg.Open.String()).Changed {
			pkg.Log(pkg.Open)
		} else if cmd.Flag(pkg.Queued.String()).Changed {
			pkg.Log(pkg.Queued)
		} else if cmd.Flag(pkg.Done.String()).Changed {
			pkg.Log(pkg.Done)
		} else {
			pkg.Log(pkg.Default)
		}
	},
}

var tasksAll bool
var tasksOpen bool
var tasksQueued bool
var tasksDone bool

func init() {
	rootCmd.AddCommand(tasksCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tasksCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tasksCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	tasksCmd.Flags().BoolVarP(&tasksAll, pkg.All.String(), "a", false, "Show all tasks: open, queued, done")
	tasksCmd.Flags().BoolVarP(&tasksOpen, pkg.Open.String(), "o", false, "Show only open tasks")
	tasksCmd.Flags().BoolVarP(&tasksQueued, pkg.Queued.String(), "q", false, "Show only queued tasks")
	tasksCmd.Flags().BoolVarP(&tasksDone, pkg.Done.String(), "d", false, "Show only done tasks")
}
