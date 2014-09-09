package main

import (
	"fmt"

	"os"
	"path/filepath"
	"testing"
)

func TestAuthenticateUser(t *testing.T) {
}

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

func TestExecGitHttpBackend(t *testing.T) {
	cwd, _ := os.Getwd()
	prependEnvPath(filepath.Join(cwd, "fixtures", "git-http-backend"))

	// execGitHttpBackend(...)
}
