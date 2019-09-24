package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TasksFilter int

const (
	Default TasksFilter = iota
	All
	Open
	Queued
	Done
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Default-0]
	_ = x[All-1]
	_ = x[Open-2]
	_ = x[Queued-3]
	_ = x[Done-4]
}

const _TasksFilter_name = "DefaultAllOpenQueuedDone"

var _TasksFilter_index = [...]uint8{0, 7, 10, 14, 20, 24}

func (i TasksFilter) String() string {
	if i < 0 || i >= TasksFilter(len(_TasksFilter_index)-1) {
		return "TasksFilter(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return strings.ToLower(_TasksFilter_name[_TasksFilter_index[i]:_TasksFilter_index[i+1]])
}

func Log(filter TasksFilter) {

	db := getConfDB()
	p := getCurrentProject(db)

	var tasks []Task
	var err error
	switch filter {
	case Default:
		tasks, err = getOpenQueuedTasks(db)
	case All:
		tasks, err = getAllProjectTasks(db)
	case Open:
		tasks, err = getOpenTasks(db)
	case Queued:
		tasks, err = getQueuedTasks(db)
	case Done:
		tasks, err = getDoneTasks(db)
	}

	db.Close()
	if err != nil {
		fmt.Printf("error loading tasks: %v\n", err)
		os.Exit(1)
	}

taskList:
	selectTaskLabel := "Hit enter to edit, commit, swap or delete a task. Selection: "

	task, err := promptSelectTask(selectTaskLabel, p.CurrentTaskID, sortOpenToTop(p.CurrentTaskID, tasks))
	if err != nil {
		os.Exit(1)
	}

	selectedAction := Cancel
	if p.CurrentTaskID == task.ID {
		selectedAction = promptTaskMenu(openTaskActions)
	} else if task.Done == 0 {
		selectedAction = promptTaskMenu(queuedTaskActions)
	} else {
		selectedAction = promptTaskMenu(doneTaskActions)
	}

	switch selectedAction {
	case Cancel:
		goto taskList
	case Edit:
		err = editSelectedTask(task)
		if err != nil {
			os.Exit(1)
		}
	case Swap:
		err = swapSelectedTask(task)
		if err != nil {
			os.Exit(1)
		}

	case Commit:
		err = selectAndCommitTask()
		if err != nil {
			os.Exit(1)
		}
	case Delete:
		err = deleteSelectedTask(task)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}
func sortOpenToTop(currentTaskID string, tasks []Task) []Task {
	var sorted []Task
	var remaining []Task

	for _, t := range tasks {
		if t.ID == currentTaskID {
			sorted = append(sorted, t)
		} else {
			remaining = append(remaining, t)
		}
	}

	sorted = append(sorted, remaining...)
	return sorted
}
