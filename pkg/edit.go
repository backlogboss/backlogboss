package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/asdine/storm"
)

func EditTask() {

	db := getConfDB()
	defer db.Close()

	task, err := getCurrentTask(db)
	if err != nil {
		return
	}

	err = editTask(db, task)
	if err != nil {
		fmt.Println("Edit failed.")
		return
	}
}

func editSelectedTask(task Task) error {
	db := getConfDB()
	defer db.Close()

	err := editTask(db, task)
	if err != nil {
		fmt.Println("Edit failed.")
		return err
	}
	return nil
}

func editTask(db *storm.DB, task Task) error {
	desc, err := captureInputFromEditor(
		task.Description,
		getPreferredEditorFromEnvironment,
	)
	if err != nil {
		return err
	}

	err = updateTaskDescription(db, task.ID, desc)
	if err != nil {
		return err
	}
	return nil
}

func EditAppend(message string) {
	db := getConfDB()
	defer db.Close()

	task, err := getCurrentTask(db)
	if err != nil {
		return
	}

	var desc []byte
	desc = append(desc, task.Description...)
	desc = append(desc, fmt.Sprintf("\n%s", message)...)

	err = updateTaskDescription(db, task.ID, desc)
	if err != nil {
		return
	}
}

func EditTitle(newTitle string) {
	db := getConfDB()
	defer db.Close()

	if len(newTitle) < 10 || len(newTitle) > 100 {
		fmt.Println("Title length required to be between 10 and 100 character")
		os.Exit(1)
		return
	}

	err := updateCurrentTaskTitle(db, newTitle)
	if err != nil {
		os.Exit(1)
		return
	}
}

// https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

// DefaultEditor is vim because we're adults ;)
const defaultEditor = "vim"

// PreferredEditorResolver is a function that returns an editor that the user
// prefers to use, such as the configured `$EDITOR` environment variable.
type preferredEditorResolver func() string

// GetPreferredEditorFromEnvironment returns the user's editor as defined by the
// `$EDITOR` environment variable, or the `DefaultEditor` if it is not set.
func getPreferredEditorFromEnvironment() string {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return defaultEditor
	}

	return editor
}

func resolveEditorArguments(executable string, filename string) []string {
	args := []string{filename}

	if strings.Contains(executable, "Visual Studio Code.app") {
		args = append([]string{"--wait"}, args...)
	}

	// Other common editors

	return args
}

// OpenFileInEditor opens filename in a text editor.
func openFileInEditor(filename string, resolveEditor preferredEditorResolver) error {
	// Get the full executable path for the editor.
	executable, err := exec.LookPath(resolveEditor())
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, resolveEditorArguments(executable, filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CaptureInputFromEditor opens a temporary file in a text editor and returns
// the written bytes on success or an error on failure. It handles deletion
// of the temporary file behind the scenes.
func captureInputFromEditor(initialData []byte, resolveEditor preferredEditorResolver) ([]byte, error) {
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return []byte{}, err
	}

	filename := file.Name()

	file.Write(initialData)

	// Defer removal of the temporary file in case any of the next steps fail.
	defer os.Remove(filename)

	if err = file.Close(); err != nil {
		return []byte{}, err
	}

	if err = openFileInEditor(filename, resolveEditor); err != nil {
		return []byte{}, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}
