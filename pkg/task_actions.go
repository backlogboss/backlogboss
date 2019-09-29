package pkg

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type TaskAction int

const (
	Edit TaskAction = iota
	Commit
	Swap
	Score
	Delete
	Cancel
)

var (
	openTaskActions   = []TaskAction{Edit, Commit}
	queuedTaskActions = []TaskAction{Edit, Swap, Score, Delete}
	doneTaskActions   = []TaskAction{Edit}
)

func promptTaskMenu(actions []TaskAction) TaskAction {
	action := Cancel

	actions = append([]TaskAction{Cancel}, actions...)

	funcMap := promptui.FuncMap
	funcMap["styledAction"] = func(action TaskAction, focus string) string {
		var title string
		focusTitleColor := cyan
		cursor := ""

		switch focus {

		case "active":
			focusTitleColor = boldMagenta
			cursor = "\U0001F449"
			if action == Cancel {
				cursor = "\U0001F448"
			}
		case "inactive":

			cursor = ""
		case "selected":
			focusTitleColor = green
			cursor = "\U0001F449"
		}

		actionTitle := action.String()
		if action == Cancel {
			actionTitle = "Back"
		}

		title = fmt.Sprintf("%s %s",
			cursor,
			focusTitleColor.Sprint(actionTitle),
		)

		return title
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }} ?",
		Active:   selectActionActive,
		Inactive: selectActionInactive,
		Selected: selectActionSelected,
		Details:  "",
	}

	searcher := func(input string, index int) bool {
		action := actions[index]
		name := strings.Replace(strings.ToLower(action.String()), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select operation on task ",
		Items:     actions,
		Templates: templates,
		Size:      5,
		Searcher:  searcher,
	}
	// clear the screen
	print("\033[H\033[2J")

	i, _, err := prompt.Run()
	if err != nil {
		return action
	}

	return actions[i]
}
