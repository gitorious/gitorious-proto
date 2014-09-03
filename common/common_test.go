package common

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetRealRepoPath(t *testing.T) {
	var tests = []struct {
		success       bool
		username      string
		expectedPath  string
		expectedError bool
	}{
		{true, "sickill", "sickill@THE/PATH.GIT", false},
		{false, "sickill", "", true},
	}

	var success bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if success {
			username := r.URL.Query().Get("username")
			realPath := strings.ToUpper(r.URL.Query().Get("path"))
			fmt.Fprint(w, fmt.Sprintf("%v@%v\n", username, realPath))
		} else {
			http.Error(w, "nope", http.StatusForbidden)
		}
	}))
	defer ts.Close()

	for _, test := range tests {
		success = test.success

		realPath, err := GetRealRepoPath("the/path.git", test.username, ts.URL)

		if realPath != test.expectedPath {
			t.Errorf("expected realPath to eq \"%v\", got \"%v\"", test.expectedPath, realPath)
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
		path, err := GetFullRepoPath(test.repoPath, "fixtures/repos")

		if path != test.expectedPath {
			t.Errorf("expected path %v, got %v (%v)", test.expectedPath, path, test)
		}

		var errorHappened bool
		if err != nil {
			errorHappened = true
		}
		if errorHappened != test.expectedError {
			t.Errorf("expected error %v, got \"%v\" (%v)", test.expectedError, err, test)
		}
	}
}
