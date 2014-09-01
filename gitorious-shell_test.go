package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitCommand(t *testing.T) {
	var tests = []struct {
		fullCommand     string
		expectedCommand string
		expectedPath    string
		expectedError   bool
	}{
		{"git-upload-pack 'the/path.git'", "git-upload-pack", "the/path.git", false},
		{"git upload-pack  'the/path.git'", "git upload-pack", "the/path.git", false},
		{"git-receive-pack 'the/path.git'", "git-receive-pack", "the/path.git", false},
		{"git receive-pack 'the/path.git'", "git receive-pack", "the/path.git", false},
		{"git-upload-archive 'the/path.git'", "git-upload-archive", "the/path.git", false},
		{"git upload-archive 'the/path.git'", "git upload-archive", "the/path.git", false},
		{"git update-ref 'the/path.git'", "", "", true},
		{"git-upload-pack the/path.git", "", "", true},
		{"cvs server 'the/path.git'", "", "", true},
	}

	for _, test := range tests {
		command, path, err := parseGitCommand(test.fullCommand)

		if command != test.expectedCommand {
			t.Errorf("expected command %v, got %v (%v)", test.expectedCommand, command, test)
		}

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

func TestGetFullRepoPath(t *testing.T) {
	var tests = []struct {
		repoPath      string
		expectedPath  string
		expectedError bool
	}{
		{"repo-with-hook.git", "fixtures/repos/repo-with-hook.git", false},
		{"repo-with-not-executable-hook.git", "", true},
		{"repo-without-hook.git", "", true},
		{"non-existent.git", "", true},
	}

	for _, test := range tests {
		path, err := getFullRepoPath(test.repoPath, "fixtures/repos")

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

func TestFormatGitShellCommand(t *testing.T) {
	expected := "git upload-pack '/repo/path.git'"
	actual := formatGitShellCommand("git upload-pack", "/repo/path.git")

	if actual != expected {
		t.Errorf("expected full command \"%v\", got \"%v\"", expected, actual)
	}
}

func prependEnvPath(path string) {
	oldPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%v:%v", path, oldPath)
	os.Setenv("PATH", newPath)
}

func TestExecGitShell(t *testing.T) {
	cwd, _ := os.Getwd()

	prependEnvPath(filepath.Join(cwd, "fixtures", "git-shell-success"))

	stderr, err := execGitShell("git-upload-pack '/the/repo.git'")
	if stderr != "" || err != nil {
		t.Errorf("didn't expect output on stderr nor error")
	}

	prependEnvPath(filepath.Join(cwd, "fixtures", "git-shell-failure"))

	expectedStderr := "-c git-upload-pack '/the/repo.git'"
	stderr, err = execGitShell("git-upload-pack '/the/repo.git'")
	if stderr != expectedStderr || err == nil {
		t.Errorf("expected output \"%v\" on stderr or error", expectedStderr)
	}
}