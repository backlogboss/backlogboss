package pkg

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func deleteSelectedTask(task Task) error {
	db := getConfDB()
	defer db.Close()

	if task.Done != 0 {
		return fmt.Errorf("only queued tasks can be deleted")
	}

	tx, err := db.Begin(true)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	err = tx.UpdateField(&task, "DeletedAt", &now)
	if err != nil {
		return errors.WithStack(err)
	}

	err = tx.UpdateField(&task, "UpdatedAt", time.Now())
	if err != nil {
		return errors.WithStack(err)
	}

	tx.Commit()

	fmt.Printf("Deleted task: %s", task.Title)
	return nil
}
