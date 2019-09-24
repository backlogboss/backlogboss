package pkg

import (
	"fmt"
	"os"
)

const (
	welcomeTaskID = "welcome_task_id"
	version       = "0.1"
)

var welcomeTask = Task{
	ID:          welcomeTaskID,
	Title:       "Look at Me, I'm The Captain Now.",
	Description: []byte(short),
}

// Init a new project in the current git directory
func Init() {
	db := getConfDB()
	defer db.Close()

	if currentProjectExists(db) {
		fmt.Printf(`Project '%s' is already intitialsed in the current directory.`,
			getCurrentProjectName())
		return
	}
	initProject(db)
}

func getConfDir() string {
	var err error
	home := os.Getenv("BACKLOGBOSS_CONF")
	if home == "" {
		home, err = os.UserHomeDir()
		if err != nil {
			fmt.Println("Please set BACKLOGBOSS_CONF to continue.")
			os.Exit(1)
		}
	}

	confDir := fmt.Sprintf("%s/.backlogboss", home)
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		err = os.MkdirAll(confDir, os.ModePerm)
		if err != nil {
			fmt.Printf("error creating backlogboss conf dir %v %v", confDir, err)
			os.Exit(1)
		}
	}

	return confDir
}

func exitProjectNotFound() {
	fmt.Println("Project not found in the current working directory.")
	fmt.Println("To initialize a project: $ backlogboss init")
	os.Exit(1)
}

func indexDB() {
	db := getConfDB()
	defer db.Close()

	err := db.Init(&User{})
	if err != nil {
		fmt.Println("err indexing user")
		os.Exit(1)
	}
	err = db.Init(&TokenResponse{})
	if err != nil {
		fmt.Println("err indexing token")
		os.Exit(1)
	}
	err = db.Init(&Project{})
	if err != nil {
		fmt.Println("err indexing project")
		os.Exit(1)
	}
	err = db.Init(&Task{})
	if err != nil {
		fmt.Println("err indexing task")
		os.Exit(1)
	}
}
