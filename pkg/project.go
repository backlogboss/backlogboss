package pkg

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/magiconair/properties"

	"github.com/asdine/storm"
	"github.com/google/uuid"
)

// Project represents a single intialized project
type Project struct {
	ID        string    `storm:"id" json:"id"`
	CreatedAt time.Time `storm:"index" json:"created_at"`
	UpdatedAt time.Time `storm:"index" json:"updated_at"`

	UserID        string `storm:"index" json:"user_id"`
	CLIVersion    string `json:"cli_version"`
	ServerVersion string `json:"server_version"`

	Name          string `storm:"unique" json:"name"`
	CurrentTaskID string `json:"current_task_id"`
}

func getCurrentProjectName() string {
	var name string
	// local project properties file
	props, err := properties.LoadFile(".backlogboss", properties.UTF8)
	if err == nil {
		name = props.GetString("project_name", "")
		if name != "" {
			return name
		}
	}

	name = os.Getenv("BACKLOGBOSS_PROJECT_NAME")
	if name != "" {
		return name
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting project name %v", err)
		fmt.Printf(`Please set 'project_name=myproject' in a .backlogboss file in the current dir. OR`)
		fmt.Printf("Please set local env var: BACKLOGBOSS_PROJECT_NAME=myproject.")
		fmt.Printf("A project name is unique.")
		os.Exit(1)
	}

	var ss []string
	if runtime.GOOS == "windows" {
		ss = strings.Split(dir, "\\")
	} else {
		ss = strings.Split(dir, "/")
	}

	name = ss[len(ss)-1]

	return name
}

func getConfDB() *storm.DB {
	db, err := storm.Open(fmt.Sprintf("%s/data", getConfDir()))
	if err != nil {
		fmt.Printf("error opening data file %v", err)
		os.Exit(1)
	}

	return db
}

func initProject(db *storm.DB) {

	project := Project{
		ID:            uuid.New().String(),
		Name:          getCurrentProjectName(),
		CurrentTaskID: welcomeTask.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CLIVersion:    version,
	}

	tx, err := db.Begin(true)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.Save(&project)
	if err != nil {
		return
	}

	welcomeTask.ID = fmt.Sprintf("%s-%s", welcomeTask.ID, project.ID)
	welcomeTask.ProjectID = project.ID
	welcomeTask.UserID = project.UserID
	now := time.Now()
	welcomeTask.UpdatedAt = now
	welcomeTask.CreatedAt = now

	err = tx.Save(&welcomeTask)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	fmt.Printf("\nInitialised new project ...\n\n")
	printTaskFull(welcomeTask)
	return
}

func getCurrentProject(db *storm.DB) Project {
	name := getCurrentProjectName()

	var project Project
	err := db.One("Name", name, &project)
	if err != nil {
		if err == storm.ErrNotFound {
			exitProjectNotFound()
		}
		fmt.Printf("unexpected error %v", err)
		os.Exit(1)
	}

	return project
}

func currentProjectExists(db *storm.DB) bool {
	name := getCurrentProjectName()
	var p Project
	err := db.One("Name", name, &p)
	if err == nil && p.ID != "" {
		return true
	}
	return false
}

func getProjectByID(db *storm.DB, id string) (Project, error) {
	var project Project
	err := db.One("ID", id, &project)
	if err != nil {
		return project, err
	}
	return project, nil
}

func getAllProjects(db *storm.DB) ([]Project, error) {
	var projects []Project
	err := db.AllByIndex("UpdatedAt", &projects)
	if err != nil {
		return projects, err
	}
	return projects, nil
}
