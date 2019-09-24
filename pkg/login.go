package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"

	"github.com/skratchdot/open-golang/open"

	"github.com/asdine/storm"
)

var (
	host        = getHost()
	identityAPI = fmt.Sprintf("%s/identity", host)
	baseAPI     = fmt.Sprintf("%s/api", host)
	healthAPI   = fmt.Sprintf("%s/health", host)
	aud         = "api.backlogboss.xyz"
)

func getHost() string {
	if os.Getenv("BACKLOGBOSS_SERVER") != "" {
		return os.Getenv("BACKLOGBOSS_SERVER")
	}
	return "http://localhost:5454"
}

type TokenResponse struct {
	ID           string `storm:"unique" json:"id"`
	Aud          string `json:"aud"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    time.Time
}

type User struct {
	ID                 string                 `json:"id"`
	Aud                string                 `json:"aud"`
	Role               string                 `json:"role"`
	Email              string                 `json:"email"`
	ConfirmedAt        time.Time              `json:"confirmed_at"`
	ConfirmationSentAt time.Time              `json:"confirmation_sent_at"`
	RecoverySentAt     time.Time              `json:"recovery_sent_at"`
	AppMetadata        map[string]interface{} `json:"app_metadata"`
	UserMetadata       map[string]interface{} `json:"user_metadata"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

func Login() {
	indexDB()
	db := getConfDB()
	defer db.Close()

	token, err := getToken(db)
	if err == nil {
		user, err := getUserOffline(db)
		if err == nil && user != nil {
			fmt.Printf("%s Already logged in: %v\nTo logout: $ backlogboss logout \n", thumbsUp, user.Email)
			return
		}
	}

	username, password, err := promptLogin()
	if err != nil {
		os.Exit(1)
	}

	token, err = fetchToken(db, passwordGrant(username, password))
	if err != nil {
		fmt.Printf("error logging in: %v ", err)
		os.Exit(1)
	}

	user, err := fetchUser(db, token.AccessToken)
	if err != nil || user == nil {
		fmt.Printf("error creating logged in user: %v ", err)
		os.Exit(1)
	}

	tasks, err := getAllTasks(db)
	if err != nil {
		// TODO: handle error
		os.Exit(0)
	}

	for _, task := range tasks {
		if task.UserID == "" {
			task.UserID = user.ID
			if err := db.Save(&task); err != nil {
				// TODO: handle error
				os.Exit(0)
			}
		}
	}

	projects, err := getAllProjects(db)
	if err != nil {
		// TODO: handle error
		os.Exit(0)
	}

	for _, project := range projects {
		if project.UserID == "" {
			project.UserID = user.ID
			if err := db.Save(&project); err != nil {
				// TODO: handle error
				os.Exit(0)
			}
		}
	}

	fmt.Printf("%s User %v logged in\n", thumbsUp, user.Email)
}

func promptLogin() (string, string, error) {
	prompt := promptui.Prompt{
		Label: "Username",
	}

	username, err := prompt.Run()
	if err != nil {
		return "", "", err
	}

	prompt = promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	password, err := prompt.Run()
	if err != nil {
		return "", "", err
	}

	return username, password, nil
}

func Logout() {

	db := getConfDB()
	defer db.Close()

	err := clearUserData(db)
	if err != nil {
		fmt.Errorf("erorr logging out %w", err)
		return
	}

	fmt.Printf("%s You are logged out.\n", thumbsDown)
	fmt.Println("To login: $ backlogboss login")
}

func clearUserData(db *storm.DB) error {
	err := db.Drop(&User{})
	if err != nil {
		return err
	}
	err = db.Drop(&TokenResponse{})
	if err != nil {
		return err
	}
	return nil
}

func Signup() {
	fmt.Println("Sign up at: https://backlogboss.xyz/signup")
	open.Run("https://backlogboss.xyz/signup")
}

func Web() {
	open.Run("https://backlogboss.xyz/dashboard")
}

func fetchUser(db *storm.DB, token string) (*User, error) {
	req, err := http.NewRequest("GET", identityAPI+"/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-JWT-AUD", aud)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := defaultClient().Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("fetch user failed with status %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	err = db.Update(&user)
	if err != nil {
		if err == storm.ErrNotFound {
			err = db.Save(&user)
			if err != nil {
				return nil, err
			}
			return &user, nil
		}
		return nil, err
	}

	return &user, nil
}

func passwordGrant(username, password string) string {
	params := url.Values{}
	params.Add("grant_type", "password")
	params.Add("username", username)
	params.Add("password", password)
	return params.Encode()
}

func refreshGrant(token string) string {
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", token)
	return params.Encode()
}

func getUserRemote(db *storm.DB) (*User, error) {
	token, err := getToken(db)
	if err != nil {
		return nil, err
	}

	return fetchUser(db, token.AccessToken)
}

func getUserOffline(db *storm.DB) (*User, error) {
	var users []User
	err := db.All(&users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("unexpected. too many users")

	}
	return &users[0], nil
}

func getToken(db *storm.DB) (*TokenResponse, error) {
	var tokens []TokenResponse
	err := db.All(&tokens)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	if len(tokens) > 1 {
		return nil, fmt.Errorf("unexpected. too many users")
	}

	token := tokens[0]
	if token.Aud != getHost() {
		return nil, fmt.Errorf("aud is different")
	}

	if token.ExpiresAt.Before(time.Now()) {
		return fetchToken(db, refreshGrant(token.RefreshToken))
	}

	return &token, nil
}

func fetchToken(db *storm.DB, grant string) (*TokenResponse, error) {
	payload := strings.NewReader(grant)
	req, err := http.NewRequest("POST", identityAPI+"/token", payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := defaultClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("fetch token failed with status %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tokenRes TokenResponse
	err = json.Unmarshal(body, &tokenRes)
	if err != nil {
		return nil, err
	}

	tokenRes.ExpiresAt = time.Now().Add(time.Second * time.Duration(tokenRes.ExpiresIn))
	tokenRes.ID = "token-id"
	tokenRes.Aud = getHost()

	err = db.Save(&tokenRes)
	if err != nil {
		return nil, err
	}

	return &tokenRes, nil
}

func okapi(token string) error {
	req, err := http.NewRequest("GET", baseAPI+"/ok", nil)
	if err != nil {
		return err
	}

	req.Header.Add("X-JWT-AUD", aud)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := defaultClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func health() bool {

	req, err := http.NewRequest("GET", healthAPI, nil)
	if err != nil {
		return false
	}

	res, err := defaultClient().Do(req)
	if err != nil {
		return false
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}
	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false
	}

	v, ok := result["status"]
	if !ok {
		return false
	}
	if v != "UP" {
		return false
	}

	return true

}
