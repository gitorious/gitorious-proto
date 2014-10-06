package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type RepoConfig struct {
	RepositoryId int    `json:"repository_id"`
	FullPath     string `json:"full_path"`

	SshCloneUrl  string `json:"ssh_clone_url"`
	HttpCloneUrl string `json:"http_clone_url"`
	GitCloneUrl  string `json:"git_clone_url"`

	CustomPreReceivePath  string `json:"custom_pre_receive_path"`
	CustomPostReceivePath string `json:"custom_post_receive_path"`
	CustomUpdatePath      string `json:"custom_update_path"`
}

type User struct {
	Username string `json:"username"`
}

type InternalApi interface {
	GetRepoConfig(string, string) (*RepoConfig, error)
	AuthenticateUser(string, string) (*User, error)
}

type HttpError struct {
	Url        *url.URL
	StatusCode int
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("got HTTP status %v for %v", e.StatusCode, e.Url)
}

type GitoriousInternalApi struct {
	ApiUrl string
}

func (a *GitoriousInternalApi) GetRepoConfig(repoPath, username string) (*RepoConfig, error) {
	u, err := url.Parse(a.ApiUrl + "/repo-config")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("repo_path", repoPath)
	q.Set("username", username)
	u.RawQuery = q.Encode()

	var repoConfig RepoConfig

	if err := a.getJson(u, &repoConfig); err != nil {
		return nil, err
	}

	return &repoConfig, nil
}

func (a *GitoriousInternalApi) AuthenticateUser(username, password string) (*User, error) {
	u, err := url.Parse(a.ApiUrl + "/authenticate")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("username", username)
	q.Set("password", password)
	u.RawQuery = q.Encode()

	var user User

	if err := a.getJson(u, &user); err != nil {
		if httpErr, ok := err.(*HttpError); ok {
			if httpErr.StatusCode == 401 {
				return nil, nil
			}
		}

		return nil, err
	}

	return &user, nil
}

func (a *GitoriousInternalApi) getJson(u *url.URL, target interface{}) error {
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	request.Header.Add("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return &HttpError{u, response.StatusCode}
	}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(target)
	if err != nil {
		return err
	}

	return nil
}
