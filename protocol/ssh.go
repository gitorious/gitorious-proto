package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

type SshProtocolHandler struct {
	username   string
	sshCommand string
}

func (h *SshProtocolHandler) GetUsername() (string, error) {
	return h.username, nil
}

var gitCommandRegexp = regexp.MustCompile("^(git(-|\\s)(receive-pack|upload-pack|upload-archive))\\s+'/?([^']+)'$")

func (h *SshProtocolHandler) ParseCommand() (string, string, error) {
	matches := gitCommandRegexp.FindStringSubmatch(h.sshCommand)
	if matches == nil {
		return "", "", errors.New(fmt.Sprintf(`invalid git-shell command "%v"`, h.sshCommand))
	}

	return matches[4], matches[1], nil
}

func (h *SshProtocolHandler) RunProxy(repoConfig RepositoryConfiguration, command, username string) (string, error) {
	syscall.Umask(0022) // set umask for pushes

	env := os.Environ()
	env = append(env, "GITORIOUS_PROTO=ssh")
	env = append(env, "GITORIOUS_USER="+username) // utilized by hooks

	var stderrBuf bytes.Buffer
	cmd := exec.Command("git-shell", "-c", command) // +path from repoConfig and reposRoot
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return strings.Trim(stderrBuf.String(), " \n"), err
	}

	return "", nil
}
