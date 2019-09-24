package pkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

type TaskAction int

const (
	Edit TaskAction = iota
	Commit
	Swap
	Delete
	Cancel
)

var (
	openTaskActions   = []TaskAction{Edit, Commit}
	queuedTaskActions = []TaskAction{Edit, Swap, Delete}
	doneTaskActions   = []TaskAction{Edit}
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Edit-0]
	_ = x[Commit-1]
	_ = x[Swap-2]
	_ = x[Delete-3]
	_ = x[Cancel-4]
}

const _TaskAction_name = "EditCommitSwapDeleteCancel"

var _TaskAction_index = [...]uint8{0, 4, 10, 14, 20, 26}

func (i TaskAction) String() string {
	if i < 0 || i >= TaskAction(len(_TaskAction_index)-1) {
		return "TaskAction(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TaskAction_name[_TaskAction_index[i]:_TaskAction_index[i+1]]
}

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
