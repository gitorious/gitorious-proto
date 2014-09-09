package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type GitoriousPublicApi struct {
	apiUrl   string
	username string
	password string
}

func NewGitoriousPublicApi(apiUrl, username, password string) *GitoriousPublicApi {
	return &GitoriousPublicApi{apiUrl: apiUrl, username: username, password: password}
}

func (a *GitoriousPublicApi) GetUserInfo() (*User, error) {
	var user User
	err := a.getJson("/user", &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *GitoriousPublicApi) getJson(path string, target interface{}) error {
	url := a.apiUrl + path
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	request.SetBasicAuth(a.username, a.password)
	request.Header.Add("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("got status %v from %v", response.StatusCode, url))
	}

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(target)
	if err != nil {
		return err
	}

	return nil
}
