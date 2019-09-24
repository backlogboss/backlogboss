package pkg

import (
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
)

type statusData struct {
	Updated     string
	Created     string
	Title       string
	Description string
}

type infoData struct {
	Server         string
	User           string
	Projects       []Project
	CurrentProject Project
}

func Status() {
	db := getConfDB()
	defer db.Close()

	task, err := getCurrentTask(db)
	if err != nil {
		os.Exit(1)
	}

	printStyledOutput(statusTmpl, statusData{
		Updated:     humanize.Time(task.UpdatedAt),
		Created:     humanize.Time(task.CreatedAt),
		Title:       task.Title,
		Description: string(task.Description),
	})
}

func Info() {
	db := getConfDB()
	defer db.Close()
	userEmail := "not logged in"

	if health() {
		token, err := getToken(db)
		if err == nil {
			okapi(token.AccessToken)
			user, errUser := getUserOffline(db)
			if errUser == nil && user != nil {
				userEmail = user.Email
			}
		}
	} else {
		fmt.Printf("Server %v is unreachable", getHost())
	}

	projects, err := getAllProjects(db)
	if err != nil {
		os.Exit(1)
	}

	p := getCurrentProject(db)

	printStyledOutput(infoTmpl, infoData{
		Server:         getHost(),
		User:           userEmail,
		Projects:       projects,
		CurrentProject: p,
	})
}

func printTask(task Task, currentTaskID string) {
	if task.ID == currentTaskID {
		fmt.Printf("\n(Current) Title: %v\n", task.Title)
	} else {
		fmt.Printf("\nTitle: %v\n", task.Title)
	}

	fmt.Printf("Commit Message: %v\n", string(task.CommitMessage))
	fmt.Printf("Updated: %v  Created: %v \n", humanize.Time(task.UpdatedAt), humanize.Time(task.CreatedAt))
}

func printTaskFull(task Task) {
	fmt.Println(" ---------------------------------")
	fmt.Println(string(task.Description))
}
