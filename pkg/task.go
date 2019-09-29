package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/liamg/tml"

	"github.com/fatih/color"

	"github.com/chzyer/readline"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

// stderr implements an io.WriteCloser that skips the terminal bell character
// (ASCII code 7), and writes the rest to os.Stderr. It's used to replace
// readline.Stdout, that is the package used by promptui to display the prompts.
type stderr struct{}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (s *stderr) Write(b []byte) (int, error) {
	if len(b) == 1 && b[0] == readline.CharBell {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (s *stderr) Close() error {
	return os.Stderr.Close()
}

func init() {
	readline.Stdout = &stderr{}
}

// Task ...
type Task struct {
	ID            string     `storm:"id" json:"id"`
	UserID        string     `json:"user_id"`
	ProjectID     string     `storm:"index" json:"project_id"`
	Title         string     `json:"title"`
	Description   []byte     `json:"description"`
	CommitMessage []byte     `json:"commit_message"`
	Done          int        `json:"done"`
	Score         int64      `json:"score"`
	CreatedAt     time.Time  `storm:"index" json:"created_at"`
	UpdatedAt     time.Time  `storm:"index" json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

func getCurrentTask(db *storm.DB) (Task, error) {
	project := getCurrentProject(db)
	task, err := getTaskByID(db, project.CurrentTaskID)
	if err != nil {
		return task, err
	}
	return task, nil
}

func getTaskByID(db *storm.DB, id string) (Task, error) {
	var task Task
	err := db.One("ID", id, &task)
	if err != nil {
		return task, err
	}
	return task, nil
}

func getOpenTasks(db *storm.DB) ([]Task, error) {
	var tasks []Task
	project := getCurrentProject(db)
	task, err := getTaskByID(db, project.CurrentTaskID)
	if err != nil {
		return tasks, err
	}

	tasks = append(tasks, task)

	return tasks, nil

}

func getQueuedTasks(db *storm.DB) ([]Task, error) {
	project := getCurrentProject(db)
	var tasks []Task
	err := db.Select(
		q.Eq("DeletedAt", nil),
		q.Eq("ProjectID", project.ID),
		q.Eq("Done", 0),
		q.Not(q.Eq("ID", project.CurrentTaskID)),
	).OrderBy("UpdatedAt").Reverse().Find(&tasks)
	if err != nil {
		if err == storm.ErrNotFound {
			return tasks, nil
		}

		return nil, err
	}

	return tasks, nil
}

func getOpenQueuedTasks(db *storm.DB) ([]Task, error) {
	project := getCurrentProject(db)

	var tasks []Task
	err := db.Select(
		q.Eq("DeletedAt", nil),
		q.Eq("ProjectID", project.ID),
		q.Eq("Done", 0),
	).OrderBy("UpdatedAt").Reverse().Find(&tasks)
	if err != nil {
		if err == storm.ErrNotFound {
			return tasks, nil
		}
		return nil, err
	}

	return tasks, nil
}

func getDoneTasks(db *storm.DB) ([]Task, error) {
	project := getCurrentProject(db)

	var tasks []Task
	err := db.Select(
		q.Eq("DeletedAt", nil),
		q.Eq("ProjectID", project.ID),
		q.Eq("Done", 1),
		q.Not(q.Eq("ID", project.CurrentTaskID)),
	).OrderBy("UpdatedAt").Reverse().Find(&tasks)
	if err != nil {
		if err == storm.ErrNotFound {
			return tasks, nil
		}

		return nil, err
	}

	return tasks, nil
}

func getAllProjectTasks(db *storm.DB) ([]Task, error) {
	project := getCurrentProject(db)

	var tasks []Task
	err := db.Select(
		q.Eq("DeletedAt", nil),
		q.Eq("ProjectID", project.ID),
	).OrderBy("UpdatedAt").Reverse().Find(&tasks)
	if err != nil {
		if err == storm.ErrNotFound {
			return tasks, nil
		}
		return nil, err
	}

	return tasks, nil
}

func getAllTasks(db *storm.DB) ([]Task, error) {
	var tasks []Task
	err := db.AllByIndex("UpdatedAt", &tasks)
	if err != nil {
		if err == storm.ErrNotFound {
			return tasks, nil
		}
		return nil, err
	}

	return tasks, nil
}

func updateTaskDescription(db *storm.DB, taskID string, description []byte) error {
	err := db.Update(&Task{ID: taskID, Description: description, UpdatedAt: time.Now()})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func updateCurrentTaskTitle(db *storm.DB, title string) error {
	project := getCurrentProject(db)
	err := db.Update(&Task{ID: project.CurrentTaskID, Title: title, UpdatedAt: time.Now()})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func promptUpdateScore(task Task) error {
	label := fmt.Sprintf("Update task priority score(0-13). Current score is %v >", task.Score)
	return updateScore(task, promptScoreInput(label))
}

func updateScore(task Task, score int64) error {
	db := getConfDB()
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.UpdateField(&Task{ID: task.ID}, "Score", score)
	if err != nil {
		return errors.WithStack(err)
	}

	err = tx.UpdateField(&Task{ID: task.ID}, "UpdatedAt", time.Now())
	if err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit()
}

var errSelectCanceled = errors.New("prompt cancelled")

func promptSelectTask(label string, currentTaskID string, backlog []Task) (Task, error) {
	if len(backlog) == 0 {
		return Task{}, fmt.Errorf("no tasks found")
	}
	cancelTask := Task{ID: "cancel", Title: "Cancel"}

	var tasks []Task
	tasks = append(tasks, cancelTask)
	tasks = append(tasks, backlog...)

	taskMap := make(map[string]Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	funcMap := promptui.FuncMap
	funcMap["isMenu"] = func(id string) bool {
		if id == "cancel" {
			return true
		}
		return false
	}

	funcMap["byteToStr"] = func(str []byte) string {
		return string(str)
	}

	funcMap["status"] = func(id string) string {
		done := taskMap[id].Done == 1
		status := "Queued"
		if done {
			status = "Done"
		}
		if !done {
			status = "Queued"
		}
		if id == currentTaskID {
			status = "Open"
		}

		return status
	}

	funcMap["styledTitle"] = func(id string, focus string) string {

		var title string
		focusPrefixColor := yellow
		focusTitleColor := cyan
		cursor := ""

		done := taskMap[id].Done == 1

		switch focus {

		case "active":
			focusPrefixColor = boldMagenta
			if done {
				focusTitleColor = boldBlue
			}
			if !done {
				focusTitleColor = boldYellow
			}
			if id == currentTaskID {
				focusTitleColor = boldGreen
			}
			cursor = "\U0001F449"
			if id == "cancel" {
				cursor = "\U0001F448"
			}
		case "inactive":
			if done {
				focusPrefixColor = blue
			}
			if !done {
				focusPrefixColor = yellow
			}
			if id == currentTaskID {
				focusPrefixColor = green
			}
			cursor = ""
		case "selected":
			focusPrefixColor = green
			cursor = "\U0001F449"
		}

		if id == "cancel" {

			title = fmt.Sprintf("%s  %s",
				cursor, color.New(color.Bold, color.Underline).Sprint(taskMap[id].Title))
			return title
		}

		prefix := "[q]"
		if taskMap[id].Done == 1 {
			prefix = "[d]"
		}

		if currentTaskID == id {
			prefix = "[o]"
		}

		title = fmt.Sprintf("%s  %s %s",
			cursor,
			focusPrefixColor.Sprint(prefix),
			focusTitleColor.Sprint(taskMap[id].Title),
		)

		return title
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }} ?",
		Active:   selectTaskActive,
		Inactive: selectTaskInactive,
		Selected: selectTaskSelected,
		Details:  tml.Sprintf(selectTaskDetails),
		FuncMap:  funcMap,
	}

	searcher := func(input string, index int) bool {
		task := tasks[index]
		name := strings.Replace(strings.ToLower(task.Title), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     tasks,
		Templates: templates,
		Size:      15,
		Searcher:  searcher,
	}
	// clear the screen
	print("\033[H\033[2J")

	i, _, err := prompt.Run()
	if err != nil {
		return Task{}, err
	}

	selectedTask := tasks[i]

	if selectedTask.ID == "cancel" {
		return Task{}, errSelectCanceled
	}

	return tasks[i], nil
}

func promptScoreInput(label string) int64 {
	validate := func(input string) error {
		val, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			return errors.New("Must be an integer")
		}
		if val < 0 || val > 13 {
			return errors.New("Must be an integer between 0 & 13")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return 0
	}

	num, err := strconv.ParseInt(result, 10, 32)
	if err != nil {
		return 0
	}

	return num
}
