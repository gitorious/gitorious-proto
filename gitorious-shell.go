package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func say(s string, args ...interface{}) {
	// print message to stderr, prefixed with colored "+-" gitorious "logo" ;)
	fmt.Fprintf(os.Stderr, "\x1b[1;32m+\x1b[31m-\x1b[0m %v\n", fmt.Sprintf(s, args...))
}

func getenv(name, defaultValue string) string {
	value := os.Getenv(name)

	if value == "" {
		value = defaultValue
	}

	return value
}

func configureLogger(logfilePath, clientId string) func() {
	f, err := os.OpenFile(logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	log.SetPrefix(fmt.Sprintf("[%v] ", clientId))

	return func() { f.Close() }
}

var gitCommandRegexp = regexp.MustCompile("^(git(-|\\s)(receive-pack|upload-pack|upload-archive))\\s+'([^']+)'$")

func parseGitCommand(fullCommand string) (string, string, error) {
	matches := gitCommandRegexp.FindStringSubmatch(fullCommand)
	if matches == nil {
		return "", "", errors.New(fmt.Sprintf("invalid git-shell command \"%v\"", fullCommand))
	}

	return matches[1], matches[4], nil
}

func getRealRepoPath(repoPath, username, apiUrl string) (string, error) {
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

	return string(body), nil
}

func getFullRepoPath(repoPath, reposRootPath string) (string, error) {
	fullRepoPath := filepath.Join(reposRootPath, repoPath)

	preReceiveHookPath := filepath.Join(fullRepoPath, "hooks", "pre-receive")
	if info, err := os.Stat(preReceiveHookPath); err != nil || info.Mode()&0111 == 0 {
		return "", errors.New("pre-receive hook is missing or is not executable")
	}

	return fullRepoPath, nil
}

func formatGitShellCommand(command, repoPath string) string {
	return fmt.Sprintf("%v '%v'", command, repoPath)
}

func execGitShell(command string) (string, error) {
	var stderrBuf bytes.Buffer
	cmd := exec.Command("git-shell", "-c", command)
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return strings.Trim(stderrBuf.String(), " \n"), err
	}

	return "", nil
}

func main() {
	clientId := getenv("SSH_CLIENT", "local")
	logfilePath := getenv("LOGFILE", "/tmp/gitorious-shell.log")
	reposRootPath := getenv("REPOSITORIES", "/var/www/gitorious/repositories")
	apiUrl := getenv("API_URL", "http://localhost:8080/foo")

	closeLogger := configureLogger(logfilePath, clientId)
	defer closeLogger()

	log.Printf("client connected")

	if len(os.Args) < 2 {
		say("Error occured, please contact support")
		log.Fatalf("username argument missing, check .authorized_keys file")
	}

	username := os.Args[1]

	ssh_original_command := strings.Trim(os.Getenv("SSH_ORIGINAL_COMMAND"), " \n")
	if ssh_original_command == "" { // deny regular ssh login attempts
		say("Hey %v! Sorry, Gitorious doesn't provide shell access. Bye!", username)
		log.Fatalf("SSH_ORIGINAL_COMMAND missing, aborting...")
	}

	command, repoPath, err := parseGitCommand(ssh_original_command)
	if err != nil {
		say("Invalid git-shell command")
		log.Fatalf("%v, aborting...", err)
	}

	realRepoPath, err := getRealRepoPath(repoPath, username, apiUrl)
	if err != nil {
		say("Access denied or invalid repository path")
		log.Fatalf("%v, aborting...", err)
	}

	fullRepoPath, err := getFullRepoPath(realRepoPath, reposRootPath)
	if err != nil {
		say("Fatal error, please contact support")
		log.Fatalf("%v, aborting...", err)
	}

	gitShellCommand := formatGitShellCommand(command, fullRepoPath)
	log.Printf("invoking git-shell with \"%v\"", gitShellCommand)

	syscall.Umask(0022) // set umask for pushes

	if stderr, err := execGitShell(gitShellCommand); err != nil {
		say("Fatal error, please contact support")
		log.Printf("error occured in git-shell: %v", err)
		log.Fatalf("stderr: %v", stderr)
	}

	log.Printf("client disconnected, all ok")
}
