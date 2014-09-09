package protocol

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cgi"
	"os"
	"regexp"
	"syscall"

	"gitorious.org/gitorious/gitorious-shell/api"
)

type HttpProtocolHandler struct {
	r *http.Request
	w http.ResponseWriter
}

func (h *HttpProtocolHandler) GetUsername() (string, error) {
	var username string

	usernameOrEmail, password, ok := BasicAuth(h.r)
	if ok {
		publicApi := api.NewGitoriousPublicApi(h.publicApiUrl, usernameOrEmail, password)

		user, err := publicApi.GetUserInfo()
		if err != nil {
			return "", err
		}

		username = user.Username
	}

	return username, nil
}

var pathRegexp = regexp.MustCompile("^/(.+\\.git)(/.+)$")

func (h *HttpProtocolHandler) ParseCommand() (string, string, error) {
	matches := pathRegexp.FindStringSubmatch(h.r.URL.Path)
	if matches == nil {
		return "", "", errors.New(fmt.Sprintf(`invalid path "%v"`, path))
	}

	return matches[1], matches[2], nil
}

func (h *HttpProtocolHandler) RunProxy(repoConfig RepositoryConfiguration, command, username string) (string, error) {
	syscall.Umask(0022) // set umask for pushes

	env := os.Environ()
	env = append(env, "GITORIOUS_PROTO=http")
	env = append(env, "GITORIOUS_USER="+username) // utilized by hooks
	env = append(env, "REMOTE_USER="+username)    // enables "receive-pack" service (push) in git-http-backend when non-blank
	env = append(env, "GIT_HTTP_EXPORT_ALL=1")
	env = append(env, "PATH_TRANSLATED="+pathTranslated)

	cgiHandler := &cgi.Handler{
		Path: "/bin/sh",
		Args: []string{"-c", "git http-backend"},
		Dir:  ".",
		Env:  env,
	}

	cgiHandler.ServeHTTP(h.w, h.r)
}
