package pkg

import (
	"fmt"
	"os"
	"time"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

// Commit opens a new  ask and closes the current task.
func CommitTask() {
	db := getConfDB()
	defer db.Close()

	tasks, err := getQueuedTasks(db)
	if err != nil {
		fmt.Printf("Could not commit task. Unexpected error %v\n", err)
		os.Exit(1)
	}

	task, err := promptSelectTask("Select new task", "", tasks)
	if err != nil {
		if err == errSelectCanceled {
			os.Exit(0)
		}
		fmt.Printf("\nCould not commit task. Title must be between 10 and 100 characters")
		os.Exit(1)
	}

	err = commitCurrentTask(db, task)
	if err != nil {
		fmt.Printf("Could not commit task. Unexpected error %v", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Success! Previous task closed. New task opened.")
	fmt.Println("Run $ backlogboss status to view.")
}

func selectAndCommitTask() error {
	db := getConfDB()
	defer db.Close()

	tasks, err := getQueuedTasks(db)
	if err != nil {
		fmt.Printf("Could not commit task. Unexpected error %v\n", err)
		os.Exit(1)
	}

	// current task id can be empty
	selectedTask, err := promptSelectTask("Select new task", "", tasks)
	if err != nil {
		if err == errSelectCanceled {
			os.Exit(0)
		}
		fmt.Printf("\nCould not commit task. Title must be between 10 and 100 characters")
		os.Exit(1)
	}

	err = commitCurrentTask(db, selectedTask)
	if err != nil {
		fmt.Println("Commit failed.")
		return err
	}

	fmt.Println()
	fmt.Println("Success! Previous task closed. New task opened.")
	fmt.Println("Run $ backlogboss status to view.")
	return nil
}

func commitCurrentTask(db *storm.DB, selectedTask Task) error {
	p := getCurrentProject(db)

	tx, err := db.Begin(true)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.UpdateField(&Task{ID: p.CurrentTaskID}, "Done", 1)
	if err != nil {
		return errors.WithStack(err)
	}

	err = tx.UpdateField(&Task{ID: p.CurrentTaskID}, "UpdatedAt", time.Now())
	if err != nil {
		return errors.WithStack(err)
	}

	p.CurrentTaskID = selectedTask.ID
	p.UpdatedAt = time.Now()
	err = tx.Save(&p)
	if err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit()
}
