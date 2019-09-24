package pkg

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
)

var (
	errEmptyTitle = errors.New("title is empty")
	errLenTitle   = errors.New("title min len is 6 and max is 100")
)

func validateTitle(input string) error {
	if input == "" {
		return errEmptyTitle
	}

	if len(input) < 10 || len(input) > 100 {
		return errLenTitle
	}
	return nil
}

func promptTitle(edit bool) (string, error) {

	label := "Enter New Task Title >>"
	if edit {
		label = "Edit Task Title >>"
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     label,
		Templates: templates,
		Validate:  validateTitle,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == errEmptyTitle {
			fmt.Println("Could not commit task! Requires the title for the new task.")
			return "", err
		}

		if err == errLenTitle {
			fmt.Println("Could not commit task! Requires the title length to be between 6 and 100")
			return "", err
		}
		return "", err
	}

	return result, err
}
