package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetRealRepoPath(repoPath, username, apiUrl string) (string, error) {
	url := fmt.Sprintf("%v?username=%v&path=%v", apiUrl, username, repoPath)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("got status %v from API", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.Trim(string(body), " \n"), nil
}

func GetFullRepoPath(repoPath, reposRootPath string) (string, error) {
	fullRepoPath := filepath.Join(reposRootPath, repoPath)

	preReceiveHookPath := filepath.Join(fullRepoPath, "hooks", "pre-receive")
	fmt.Println(preReceiveHookPath)
	if info, err := os.Stat(preReceiveHookPath); err != nil || info.Mode()&0111 == 0 {
		return "", errors.New("pre-receive hook is missing or is not executable")
	}

	return fullRepoPath, nil
}
