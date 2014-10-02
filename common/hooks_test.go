package common

import "testing"

func TestPreReceiveHookExists(t *testing.T) {
	var tests = []struct {
		repoPath        string
		expectedSuccess bool
	}{
		{"repo-with-hook.git", true},
		{"repo-with-not-executable-hook.git", false},
		{"repo-without-hook.git", false},
		{"non-existent.git", false},
	}

	for _, test := range tests {
		ok := PreReceiveHookExists("fixtures/repos/" + test.repoPath)

		if ok != test.expectedSuccess {
			t.Errorf("expected success to be %v, got %v (%v)", test.expectedSuccess, ok, test)
		}
	}
}
