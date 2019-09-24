package pkg

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

func swapSelectedTask(task Task) error {
	db := getConfDB()
	defer db.Close()

	err := swapCurrentTask(db, task)
	if err != nil {
		fmt.Println("Swap failed.")
		return err
	}
	fmt.Println()
	fmt.Println("Success! Selected task is now open")
	fmt.Println("Run $ backlogboss status to view.")
	return nil
}

func swapCurrentTask(db *storm.DB, newCurrentTask Task) error {
	project := getCurrentProject(db)

	tx, err := db.Begin(true)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	project.CurrentTaskID = newCurrentTask.ID
	project.UpdatedAt = time.Now()
	// update project
	err = tx.Save(&project)
	if err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit()
}
