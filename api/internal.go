package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type RepoConfig struct {
	RealPath string `json:"real_path"`

	SshCloneUrl  string `json:"ssh_clone_url"`
	HttpCloneUrl string `json:"http_clone_url"`
	GitCloneUrl  string `json:"git_clone_url"`

	CustomPreReceivePath  string `json:"custom_pre_receive_path"`
	CustomPostReceivePath string `json:"custom_post_receive_path"`
	CustomUpdatePath      string `json:"custom_update_path"`
}

type InternalApi interface {
	GetRepoConfig(string, string) (*RepoConfig, error)
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
		return errors.New(fmt.Sprintf("got status %v from %v", response.StatusCode, u))
	}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(target)
	if err != nil {
		return err
	}

	return nil
}
