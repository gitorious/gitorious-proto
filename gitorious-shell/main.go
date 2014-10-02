package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"gitorious.org/gitorious/gitorious-proto/api"
	"gitorious.org/gitorious/gitorious-proto/common"
)

func say(s string, args ...interface{}) {
	// print message to stderr, prefixed with colored "+-" gitorious "logo" ;)
	fmt.Fprintf(os.Stderr, "\x1b[1;32m+\x1b[31m-\x1b[0m %v\n", fmt.Sprintf(s, args...))
}

var gitCommandRegexp = regexp.MustCompile("^(git(-|\\s)(receive-pack|upload-pack|upload-archive))\\s+'/?([^']+)'$")

func parseGitShellCommand(fullCommand string) (string, string, error) {
	matches := gitCommandRegexp.FindStringSubmatch(fullCommand)
	if matches == nil {
		return "", "", errors.New(fmt.Sprintf(`invalid git-shell command "%v"`, fullCommand))
	}

	return matches[1], matches[4], nil
}

func formatGitShellCommand(command, repoPath string) string {
	return fmt.Sprintf("%v '%v'", command, repoPath)
}

func getLogger(logfilePath, clientId string) common.Logger {
	var writer io.Writer

	writer, err := os.OpenFile(logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		writer = ioutil.Discard
	}

	targetLogger := log.New(writer, "", log.LstdFlags)
	return &common.SessionLogger{targetLogger, clientId}
}

func createSshEnv(username string, repoConfig *api.RepoConfig) []string {
	return common.CreateEnv("ssh", username, repoConfig)
}

func execGitShell(command string, env []string, stdin io.Reader, stdout io.Writer) (string, error) {
	cmd := exec.Command("git-shell", "-c", command)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return strings.Trim(stderrBuf.String(), " \n"), err
	}

	return "", nil
}

func main() {
	syscall.Umask(0022) // set umask for pushes

	clientId := common.Getenv("SSH_CLIENT", "local")
	logfilePath := common.Getenv("LOGFILE", "/var/log/gitorious/gitorious-shell.log")
	internalApiUrl := common.Getenv("GITORIOUS_INTERNAL_API_URL", "http://localhost:3000/api/internal")

	logger := getLogger(logfilePath, clientId)
	internalApi := &api.GitoriousInternalApi{internalApiUrl}

	logger.Printf("client connected")

	if len(os.Args) < 2 {
		say("Error occured, please contact support")
		logger.Printf("username argument missing, check .authorized_keys file")
		os.Exit(1)
	}

	username := os.Args[1]
	logger.Printf("user authenticated as %v", username)

	sshCommand := strings.Trim(os.Getenv("SSH_ORIGINAL_COMMAND"), " \n")

	if sshCommand == "" { // deny regular ssh login attempts
		say("Hey %v! Sorry, Gitorious doesn't provide shell access. Bye!", username)
		logger.Printf("SSH_ORIGINAL_COMMAND missing, aborting...")
		os.Exit(1)
	}

	logger.Printf("processing command: %v", sshCommand)

	command, repoPath, err := parseGitShellCommand(sshCommand)
	if err != nil {
		say("Invalid command")
		logger.Printf("%v, aborting...", err)
		os.Exit(1)
	}

	repoConfig, err := internalApi.GetRepoConfig(repoPath, username)
	if err != nil {
		if httpErr, ok := err.(*api.HttpError); ok {
			if httpErr.StatusCode == 403 {
				say("Access denied")
				logger.Printf("%v, aborting...", err)
				os.Exit(1)
			} else if httpErr.StatusCode == 404 {
				say("Invalid repository path")
				logger.Printf("%v, aborting...", err)
				os.Exit(1)
			}
		}

		say("Error occured, please contact support")
		logger.Printf("%v, aborting...", err)
		os.Exit(1)
	}

	logger.Printf("full repo path: %v", repoConfig.FullPath)

	if !common.PreReceiveHookExists(repoConfig.FullPath) {
		say("Error occurred, please contact support")
		logger.Printf("pre-receive hook for %v is missing or is not executable, aborting...", repoConfig.FullPath)
		os.Exit(1)
	}

	gitShellCommand := formatGitShellCommand(command, repoConfig.FullPath)
	env := createSshEnv(username, repoConfig)

	logger.Printf(`invoking git-shell with command "%v"`, gitShellCommand)

	if stderr, err := execGitShell(gitShellCommand, env, os.Stdin, os.Stdout); err != nil {
		say("Error occurred, please contact support")
		logger.Printf("error occured in git-shell: %v", err)
		logger.Printf("stderr: %v", stderr)
		os.Exit(1)
	}

	logger.Printf("done")
}
