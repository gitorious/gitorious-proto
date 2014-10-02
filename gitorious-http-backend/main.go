package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"regexp"
	"syscall"

	"gitorious.org/gitorious/gitorious-proto/api"
	"gitorious.org/gitorious/gitorious-proto/common"
)

func say(w http.ResponseWriter, status int, s string, args ...interface{}) {
	http.Error(w, fmt.Sprintf(s, args...), status)
}

func requestBasicAuth(w http.ResponseWriter, s string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Gitorious"`)
	say(w, http.StatusUnauthorized, s)
}

var pathRegexp = regexp.MustCompile("^/(.+\\.git)(/.+)$")

func parsePath(path string) (string, string, error) {
	matches := pathRegexp.FindStringSubmatch(path)
	if matches == nil {
		return "", "", errors.New(fmt.Sprintf(`invalid path "%v"`, path))
	}

	return matches[1], matches[2], nil
}

func createHttpEnv(username string, repoConfig *api.RepoConfig, translatedPath string) []string {
	env := common.CreateEnv("http", username, repoConfig)

	env = append(env, "REMOTE_USER="+username) // enables "receive-pack" service (push) in git-http-backend
	env = append(env, "GIT_HTTP_EXPORT_ALL=1") // enables clones without "git-daemon-export-ok" magic file
	env = append(env, "PATH_TRANSLATED="+translatedPath)

	return env
}

func execGitHttpBackend(env []string, w http.ResponseWriter, req *http.Request) {
	cgiHandler := &cgi.Handler{
		Path: "/bin/sh",
		Args: []string{"-c", "git http-backend"},
		Dir:  ".",
		Env:  env,
	}

	cgiHandler.ServeHTTP(w, req)
}

type Handler struct {
	logger      *log.Logger
	internalApi api.InternalApi
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := &common.SessionLogger{h.logger, req.RemoteAddr}

	logger.Printf("client connected")

	var username string

	if usernameOrEmail, password, ok := BasicAuth(req); ok {
		user, err := h.internalApi.AuthenticateUser(usernameOrEmail, password)
		if err != nil {
			say(w, http.StatusInternalServerError, "Error occured, please contact support")
			logger.Printf("%v, disconnecting...", err)
			return
		}

		if user != nil {
			username = user.Username
			logger.Printf("user authenticated as %v", username)
		} else {
			requestBasicAuth(w, "Invalid username or password")
			logger.Printf("invalid credentials, requesting basic auth, disconnecting...")
			return
		}
	}

	logger.Printf("processing request: %v", req.URL.String())

	repoPath, slug, err := parsePath(req.URL.Path)
	if err != nil {
		say(w, http.StatusBadRequest, "Invalid command")
		logger.Printf("%v, disconnecting...", err)
		return
	}

	if (req.URL.Query().Get("service") == "git-receive-pack" || slug == "/git-receive-pack") && username == "" {
		requestBasicAuth(w, "Anonymous pushing not allowed")
		logger.Printf("denying anonymous push, requesting basic auth, disconnecting...")
		return
	}

	repoConfig, err := h.internalApi.GetRepoConfig(repoPath, username)
	if err != nil {
		if httpErr, ok := err.(*api.HttpError); ok {
			if httpErr.StatusCode == 403 {
				requestBasicAuth(w, "Access denied")
				logger.Printf("%v, requesting basic auth, disconnecting...", err)
				return
			} else if httpErr.StatusCode == 404 {
				say(w, http.StatusNotFound, "Invalid repository path")
				logger.Printf("%v, disconnecting...", err)
				return
			}
		}

		say(w, http.StatusInternalServerError, "Error occured, please contact support")
		logger.Printf("%v, disconnecting...", err)
		return
	}

	logger.Printf("full repo path: %v", repoConfig.FullPath)

	if !common.PreReceiveHookExists(repoConfig.FullPath) {
		say(w, http.StatusInternalServerError, "Error occurred, please contact support")
		logger.Printf("pre-receive hook for %v is missing or is not executable, aborting...", repoConfig.FullPath)
		return
	}

	translatedPath := repoConfig.FullPath + slug
	env := createHttpEnv(username, repoConfig, translatedPath)

	logger.Printf(`invoking git-http-backend with translated path "%v"`, translatedPath)

	execGitHttpBackend(env, w, req)

	logger.Printf("done")
}

func main() {
	syscall.Umask(0022) // set umask for pushes

	var (
		internalApiUrl = flag.String("api-url", "http://localhost:3000/api/internal", "Gitorious internal API URL")
		addr           = flag.String("l", ":6000", "Address/port to listen on")
	)
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	internalApi := &api.GitoriousInternalApi{*internalApiUrl}

	logger.Printf("listening on %v", *addr)

	http.Handle("/", &Handler{logger, internalApi})
	log.Fatal(http.ListenAndServe(*addr, nil))
}
