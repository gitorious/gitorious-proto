package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitShellCommand(t *testing.T) {
	var tests = []struct {
		fullCommand     string
		expectedCommand string
		expectedPath    string
		expectedError   bool
	}{
		{"git-upload-pack 'the/path.git'", "git-upload-pack", "the/path.git", false},
		{"git-upload-pack '/the/path.git'", "git-upload-pack", "the/path.git", false},
		{"git upload-pack  'the/path.git'", "git upload-pack", "the/path.git", false},
		{"git-receive-pack 'the/path.git'", "git-receive-pack", "the/path.git", false},
		{"git receive-pack 'the/path.git'", "git receive-pack", "the/path.git", false},
		{"git-upload-archive 'the/path.git'", "git-upload-archive", "the/path.git", false},
		{"git upload-archive 'the/path.git'", "git upload-archive", "the/path.git", false},
		{"git-upload-archive ''", "", "", true},
		{"git update-ref 'the/path.git'", "", "", true},
		{"git-upload-pack the/path.git", "", "", true},
		{"cvs server 'the/path.git'", "", "", true},
	}

	for _, test := range tests {
		command, path, err := parseGitShellCommand(test.fullCommand)

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

func TestFormatGitShellCommand(t *testing.T) {
	expected := "git upload-pack '/repo/path.git'"
	actual := formatGitShellCommand("git upload-pack", "/repo/path.git")

	if actual != expected {
		t.Errorf(`expected full command "%v", got "%v"`, expected, actual)
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
	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stdin.Write([]byte("sha sha sha"))

	stderr, err := execGitShell("git-upload-pack '/the/repo.git'", []string{"JOLA=1"}, stdin, stdout)

	if stdout.String() != "-c git-upload-pack '/the/repo.git'\nJOLA=1\nsha sha sha" {
		t.Errorf("stdout output doesn't match")
	}

	if stderr != "" || err != nil {
		t.Errorf("didn't expect output on stderr nor error")
	}

	prependEnvPath(filepath.Join(cwd, "fixtures", "git-shell-failure"))
	stdin = &bytes.Buffer{}
	stdout = &bytes.Buffer{}

	stderr, err = execGitShell("git-upload-pack '/the/repo.git'", []string{"JOLA=1"}, stdin, stdout)

	if stderr != "such error" || err == nil {
		t.Errorf(`expected output on stderr doesn't match or error is nil, got "%v" on stderr`, stderr)
	}
}
