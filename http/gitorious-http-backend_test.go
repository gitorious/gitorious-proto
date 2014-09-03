package http

import (
	"fmt"
	"net/url"

	"os"
	"path/filepath"
	"testing"
)

func TestAuthenticateUser(t *testing.T) {
}

func TestParseUrl(t *testing.T) {
	var tests = []struct {
		url           string
		expectedPath  string
		expectedError bool
	}{
		{"/the/path.git/info/refs?service=git-upload-pack", "the/path.git", false},
		{"/the/path.git/info/refs?service=git-receive-pack", "the/path.git", false},
		{"/the/path.git/info/refs?service=git-upload-archive", "the/path.git", false},
		{"/the/path.git/info/refs?service=git-cvsserver", "", true},
		{"/foo/bar", "", true},
		{"/?service=git-upload-archive", "", true},
	}

	for _, test := range tests {
		uri, _ := url.Parse(test.url)
		path, err := parseUrl(uri)

		if path != test.expectedPath {
			t.Errorf("expected path %v, got %v (%v)", test.expectedPath, path, test)
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

func TestExecGitHttpBackend(t *testing.T) {
	cwd, _ := os.Getwd()
	prependEnvPath(filepath.Join(cwd, "fixtures", "git-http-backend"))

	// execGitHttpBackend(...)
}
