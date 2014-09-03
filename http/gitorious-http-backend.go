package http

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"syscall"

	"gitorious.org/gitorious/gitorious-shell/common"
)

func say(w http.ResponseWriter, s string, args ...interface{}) {
	fmt.Fprintln(w, fmt.Sprintf(s, args...))
}

func authenticateUser(username, password string) (string, error) {
	// request to API
	return username, nil
}

func parseUrl(uri *url.URL) (string, error) {
	service := uri.Query().Get("service")

	if service == "git-upload-pack" ||
		service == "git-upload-archive" ||
		service == "git-receive-pack" {
		return "the/path.git", nil
	}

	return "", nil
}

func execGitHttpBackend(w http.ResponseWriter, req *http.Request, fullRepoPath, username string) {
	log.Printf("invoking git-http-backend with PATH_TRANSLATED=\"%v\"", fullRepoPath)
	syscall.Umask(0022) // set umask for pushes

	env := os.Environ()
	env = append(env, "GITORIOUS_PROTO=http")
	env = append(env, "GITORIOUS_USER="+username) // utilized by hooks
	env = append(env, "REMOTE_USER="+username)    // enables "receive-pack" service (push) in git-http-backend
	env = append(env, "GIT_HTTP_EXPORT_ALL=1")
	env = append(env, "PATH_TRANSLATED="+fullRepoPath)

	cgiHandler := &cgi.Handler{
		Path: "/bin/sh",
		Args: []string{"-c", "git http-backend"},
		Dir:  ".",
		Env:  env,
	}

	cgiHandler.ServeHTTP(w, req)
}

type Handler struct {
	reposRootPath string
	apiUrl        string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("client connected")

	var username string
	var err error

	if usernameOrEmail, password, ok := BasicAuth(req); ok {
		username, err = authenticateUser(usernameOrEmail, password)
		if err != nil {
			say(w, "Error occured, please contact support")
			log.Printf("%v, disconnecting...", err)
			return
		}
	}
	// 	rw.Header().Set("WWW-Authenticate", "Basic realm=\"jola\"")
	// 	http.Error(rw, "jola", http.StatusUnauthorized)

	repoPath, err := parseUrl(req.URL)
	if err != nil {
		say(w, "Invalid command")
		log.Printf("%v, disconnecting...", err)
		return
	}

	realRepoPath, err := common.GetRealRepoPath(repoPath, username, h.apiUrl)
	if err != nil {
		say(w, "Access denied or invalid repository path")
		log.Printf("%v, disconnecting...", err)
		return
	}

	fullRepoPath, err := common.GetFullRepoPath(realRepoPath, h.reposRootPath)
	if err != nil {
		say(w, "Fatal error, please contact support")
		log.Printf("%v, disconnecting...", err)
		return
	}

	execGitHttpBackend(w, req, fullRepoPath, username)

	log.Printf("done")
}

func main() {
	var (
		reposRootPath = flag.String("r", ".", "Directory containing git repositories")
		apiUrl        = flag.String("u", "http://localhost:3000/foo", "...")
		addr          = flag.String("l", ":80", "Address/port to listen on")
	)
	flag.Parse()

	http.Handle("/", &Handler{*reposRootPath, *apiUrl})
	log.Fatal(http.ListenAndServe(*addr, nil))
}
