package pkg

import (
	"fmt"
	"os"
	"time"

	"github.com/asdine/storm"
	"github.com/google/uuid"
)

func Queue(title string) {
	db := getConfDB()
	defer db.Close()

	p := getCurrentProject(db)

	newTask := Task{
		ID:        uuid.New().String(),
		ProjectID: p.ID,
		UserID:    p.UserID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.Save(&newTask)
	if err != nil {
		fmt.Printf("Unable to queue task %v", err)
		os.Exit(1)
	}

	boldGreen.Printf("Queued task: %s\n", title)
}

func QueueRemove() {
	db := getConfDB()
	defer db.Close()

	id, err := selectQueuedTask(db)
	if err != nil {
		if err == errSelectCanceled {
			os.Exit(0)
		}
		os.Exit(1)
	}

	tx, err := db.Begin(true)
	if err != nil {
		fmt.Println("error removing task ", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	err = tx.UpdateField(&Task{ID: id}, "DeletedAt", &now)
	if err != nil {
		fmt.Println("error removing task ", err)
		os.Exit(1)
	}

	err = tx.UpdateField(&Task{ID: id}, "UpdatedAt", now)
	if err != nil {
		fmt.Println("error removing task ", err)
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("error removing task ", err)
		os.Exit(1)
	}

	fmt.Println(" Removed")
}

func QueueEdit() {

	db := getConfDB()
	defer db.Close()

	id, err := selectQueuedTask(db)
	if err != nil {
		if err == errSelectCanceled {
			os.Exit(0)
		}
		os.Exit(1)
	}

	task, err := getTaskByID(db, id)
	if err != nil {
		fmt.Printf("task not found %v", err)
		os.Exit(1)
	}

	err = editTask(db, task)
	if err != nil {
		fmt.Printf("edit task failed %v", err)
		os.Exit(1)
	}
}

func QueueView() {
	db := getConfDB()
	defer db.Close()

	id, err := selectQueuedTask(db)
	if err != nil {
		if err == errSelectCanceled {
			os.Exit(0)
		}
		os.Exit(1)
	}

	task, err := getTaskByID(db, id)
	if err != nil {
		fmt.Printf("task not found %v", err)
		os.Exit(1)
	}

	printTaskFull(task)
}

func selectQueuedTask(db *storm.DB) (string, error) {
	tasks, err := getQueuedTasks(db)
	if err != nil {
		return "", err
	}

	task, err := promptSelectTask("Remove task from queue", "", tasks)
	if err != nil {
		return "", err
	}

	return task.ID, nil
}
