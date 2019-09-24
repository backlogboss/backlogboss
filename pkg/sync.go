package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asdine/storm"
)

func Sync() {
	db := getConfDB()
	defer db.Close()

	token, err := getToken(db)
	if err != nil {
		//fmt.Printf("error looking up token: %v\n", err)
		yellow.Printf("%s logged out.\n", thumbsDown)
		return
	}

	user, err := getUserRemote(db)
	if err != nil {
		//fmt.Printf("error looking up user: %v\n", err)
		yellow.Printf("%s logged out.\n", thumbsDown)
		return
	}

	err = syncProjects(db, token, user.ID)
	if err != nil {
		fmt.Printf("error syncing to server: %v\n", err)
		return
	}

	err = syncTasks(db, token, user.ID)
	if err != nil {
		fmt.Printf("error syncing to server: %v\n", err)
		return
	}

}

func syncTasks(db *storm.DB, token *TokenResponse, userID string) error {
	remoteTasks, err := fetchRemoteTasks(token)
	if err != nil {
		return err
	}

	localTasks, err := getAllTasks(db)
	if err != nil {
		return err
	}

	if len(remoteTasks) == 0 {
		for _, local := range localTasks {
			err := putTask(local, userID, token.AccessToken)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// check local against remote.
	for _, remote := range remoteTasks {
		local, err := getTaskByID(db, remote.ID)
		if err != nil {
			if err == storm.ErrNotFound {
				err = db.Save(&remote)
				if err != nil {
					return err
				}
			}
			return err
		}

		// check if remote is newer
		if local.UpdatedAt.Before(remote.UpdatedAt) {
			err = db.Save(&remote)
			if err != nil {
				return err
			}
		} else if local.UpdatedAt.After(remote.UpdatedAt) {
			// if local is newer
			err = putTask(local, userID, token.AccessToken)
			if err != nil {
				return err
			}
		}
	}

	// check remote against local.
	for _, local := range localTasks {
		_, found := filterTaskByID(remoteTasks, local.ID)
		if !found {
			err = putTask(local, userID, token.AccessToken)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func filterTaskByID(tasks []Task, id string) (Task, bool) {
	var task Task
	for i := range tasks {
		task = tasks[i]
		if task.ID == id {
			return task, true
		}
	}

	return task, false
}

func fetchRemoteTasks(token *TokenResponse) ([]Task, error) {
	req, err := http.NewRequest("GET", baseAPI+"/tasks", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-JWT-AUD", aud)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Add("Content-type", "application/json")

	res, err := defaultClient().Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error fetchings tasks, statuscode %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	err = json.Unmarshal(body, &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func putTask(local Task, userID, accessToken string) error {
	local.UserID = userID

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(local)
	if err != nil {
		return fmt.Errorf("error encoding project %w", err)
	}

	req, err := http.NewRequest("PUT", baseAPI+"/tasks", b)
	if err != nil {
		return err
	}

	req.Header.Add("X-JWT-AUD", aud)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Content-type", "application/json")

	res, err := defaultClient().Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("error updating task %v, statuscode %v", local, res.StatusCode)
	}

	return nil
}

func syncProjects(db *storm.DB, token *TokenResponse, userID string) error {

	remoteProjects, err := fetchRemoteProjects(token)
	if err != nil {
		return err
	}

	localProjects, err := getAllProjects(db)
	if err != nil {
		return err
	}

	if len(remoteProjects) == 0 {
		for _, local := range localProjects {
			err := putProject(local, userID, token.AccessToken)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// check local against remote.
	for _, remote := range remoteProjects {
		local, err := getProjectByID(db, remote.ID)
		if err != nil {
			if err == storm.ErrNotFound {
				err = db.Save(&remote)
				if err != nil {
					return err
				}
			}
			return err
		}

		// check if remote is newer
		if local.UpdatedAt.Before(remote.UpdatedAt) {
			err = db.Save(&remote)
			if err != nil {
				return err
			}
		} else if local.UpdatedAt.After(remote.UpdatedAt) {
			// if local is newer
			err = putProject(local, userID, token.AccessToken)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchRemoteProjects(token *TokenResponse) ([]Project, error) {
	req, err := http.NewRequest("GET", baseAPI+"/projects", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-JWT-AUD", aud)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Add("Content-type", "application/json")

	res, err := defaultClient().Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error fetchings projects, statuscode %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var projects []Project
	err = json.Unmarshal(body, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func putProject(local Project, userID, accessToken string) error {
	local.UserID = userID

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(local)
	if err != nil {
		return fmt.Errorf("error encoding project %w", err)
	}

	req, err := http.NewRequest("PUT", baseAPI+"/projects", b)
	if err != nil {
		return err
	}

	req.Header.Add("X-JWT-AUD", aud)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Content-type", "application/json")

	res, err := defaultClient().Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("error updating project, statuscode %v", res.StatusCode)
	}

	return nil
}
