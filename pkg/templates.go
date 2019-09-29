package pkg

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/liamg/tml"
)

const long = `
This is your first task!

We will do one task at a time.
To close the current open task you need to create a new one.
Motivations: https://backlogboss.dev/motivation

You can always update the current task's log.

$ backlogboss edit # opens current task log in the text editor
$ backlogboss edit -m "Created header component" # appends message to the current task's log.
$ backlogboss edit -t "New title" # updates the title.
	
More edit optios:
	
$ backlogboss edit --last-n-commits 5 # appends last 5 git commit messages to the log.
$ backlogboss edit -todo "src/components/Header.js:5" -m "Fix nav to the top"
  # appends '@TODO(src/components/Header.js:5): Fix nav to the top' to the log.
	
To check the task history:

$ backlogboss status # displays the current open task's log
$ backlogboss log # lists all tasks.

To close the current task, you need to create a new task.

$ backlogboss commit "Title: Style header component" # closes previous and opens a new task.
$ backlogboss commit -e # opens new task log in the text editor. closes previous task.

To get help:

$ backlogboss -help # or https://backlogboss.dev/help

To receive daily nudges about tasks over email/sms, Please Signup here: https://backlogboss.dev/signup

Ok. Back to work Now.`

const short = `
Look at me... Look at me! I'm the captain now.

This is your first task.

We will do one task at a time. To close the current open task you need to create a new one.

	Motivations: https://backlogboss.dev/motivation

You can always update the current task's log.

	$ backlogboss edit -m "Created header component" # appends message to the current task's log.

To close the current task, you need to create a new task.

	$ backlogboss commit "Title: Style header component" # closes previous and opens a new task.

	$ backlogboss edit -h, backlogboss commit -h for more options.

To get help:
	$ backlogboss -h # or https://backlogboss.dev/help

	To receive daily nudges about tasks over email/sms.
	Please Signup here: https://backlogboss.dev/signup

Back to Work Now: $ backlogboss status`

const selectTaskDetails = `
{{ if (isMenu .ID) }}
<magenta>-------------------------</magenta>
{{ else }}
<magenta><bold>--------- Details ----------</bold></magenta>

<green><bold> {{ .Title }} </bold></green>

 <bg-green> <darkgrey>{{ status .ID}}</darkgrey> </bg-green>  <bg-darkgrey> Score: {{.Score}} </bg-darkgrey>  <bg-darkgrey> Updated: {{.UpdatedAt}} </bg-darkgrey>  <bg-darkgrey> Created: {{.CreatedAt}} </bg-darkgrey>

 <bg-darkgrey> Description:</bg-darkgrey>

 {{ ( .Description | byteToStr)}}
{{ end }}
`

const selectTaskActive = `{{ (styledTitle .ID "active") }}`
const selectTaskInactive = `{{ (styledTitle .ID "inactive") }}`
const selectTaskSelected = `{{ (styledTitle .ID "selected") }}`

const selectActionActive = `{{ (styledAction . "active") }}`
const selectActionInactive = `{{ (styledAction . "inactive") }}`
const selectActionSelected = `{{ (styledAction . "selected") }}`

const statusTmpl = `
 <bg-green> <darkgrey>Open Task</darkgrey> </bg-green>   <bg-darkgrey> Updated: {{.Updated}} </bg-darkgrey>  <bg-darkgrey> Created: {{.Created}} </bg-darkgrey>

 <green><bold>{{.Title}}</bold></green>
 --------------------------------------
 {{.Description}}

 (use <magenta>backlogboss edit</magenta> to update the current task)
 (use <magenta>backlogboss commit</magenta> to end the current task and begin a new one)`

const infoTmpl = `
<bg-green><darkgrey> User </darkgrey></bg-green>: {{.User}}

<bg-darkgrey> Server </bg-darkgrey>: {{.Server}}
<bg-darkgrey> Projects </bg-darkgrey>: {{range .Projects}}{{.Name}},{{end}}

<bg-magenta><lightgrey> Current Project </lightgrey></bg-magenta>:
  Name: {{.CurrentProject.Name}}
  Updated: {{.CurrentProject.UpdatedAt}}
  Created: {{.CurrentProject.CreatedAt}}`

func printStyledOutput(tmplVal string, data interface{}) {
	t := template.Must(template.New("styled").Parse(tmplVal))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = tml.Printf(tpl.String())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
