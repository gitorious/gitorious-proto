package common

import "testing"

func TestGitoriousRepositoryStore_GetFullRepoPath(t *testing.T) {
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
		store := &GitoriousRepositoryStore{"fixtures/repos"}
		path, err := store.GetFullRepoPath(test.repoPath)

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
