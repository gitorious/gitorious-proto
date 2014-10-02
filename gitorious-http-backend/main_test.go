package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"gitorious.org/gitorious/gitorious-proto/api"

	"os"
	"testing"
)

func TestParsePath(t *testing.T) {
	var tests = []struct {
		path          string
		expectedPath  string
		expectedSlug  string
		expectedError bool
	}{
		{"/the/path.git/HEAD", "the/path.git", "/HEAD", false},
		{"/the/path.git/info/refs", "the/path.git", "/info/refs", false},
		{"/the/path.git/git-upload-pack", "the/path.git", "/git-upload-pack", false},
		{"/the/path.git/git-receive-pack", "the/path.git", "/git-receive-pack", false},
		{"/the/pa\nth.git/HEAD", "", "", true},
		{"/the/path.git", "", "", true},
		{"/the/path", "", "", true},
		{"/", "", "", true},
	}

	for _, test := range tests {
		repoPath, slug, err := parsePath(test.path)

		if repoPath != test.expectedPath {
			t.Errorf("expected path %v, got %v (%v)", test.expectedPath, repoPath, test)
		}

		if slug != test.expectedSlug {
			t.Errorf("expected slug %v, got %v (%v)", test.expectedSlug, slug, test)
		}

		var errorHappened bool
		if err != nil {
			errorHappened = true
		}
		if errorHappened != test.expectedError {
			t.Errorf("expected error %v (%v)", test.expectedError, test)
		}
	}
}

func prependEnvPath(path string) {
	oldPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%v:%v", path, oldPath)
	os.Setenv("PATH", newPath)
}

type testInternalApi struct {
	FullRepoPath string
}

func (a *testInternalApi) AuthenticateUser(username, password string) (*api.User, error) {
	return &api.User{username + ":" + password}, nil
}

func (a *testInternalApi) GetRepoConfig(repoPath, username string) (*api.RepoConfig, error) {
	return &api.RepoConfig{FullPath: a.FullRepoPath}, nil
}

func TestHandler_ServeHTTP(t *testing.T) {
	cwd, _ := os.Getwd()
	prependEnvPath(filepath.Join(cwd, "fixtures", "git-http-backend"))

	logger := log.New(os.Stdout, "", log.LstdFlags)

	fullRepoPath := filepath.Join(cwd, "..", "common", "fixtures", "repos", "repo-with-hook.git")
	internalApi := &testInternalApi{fullRepoPath}

	handler := &Handler{logger, internalApi}

	req, _ := http.NewRequest("GET", "http://localhost/foo/bar.git/info/refs?service=git-upload-pack", nil)
	req.SetBasicAuth("sickill", "xxx")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %v", w.Code)
	}

	expectedBody := fmt.Sprintf(`GIT_HTTP_EXPORT_ALL=1
PATH_TRANSLATED=%v/info/refs
QUERY_STRING=service=git-upload-pack
REMOTE_USER=sickill:xxx
`, fullRepoPath)

	actualBody := w.Body.String()

	if actualBody != expectedBody {
		t.Errorf(`expected body "%v", got "%v"`, expectedBody, actualBody)
	}
}
