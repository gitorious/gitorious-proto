package common

import (
	"errors"
	"os"
	"path/filepath"
)

type RepositoryStore interface {
	GetFullRepoPath(string) (string, error)
}

type GitoriousRepositoryStore struct {
	Root string
}

func (s *GitoriousRepositoryStore) GetFullRepoPath(path string) (string, error) {
	fullRepoPath := filepath.Join(s.Root, path)
	preReceiveHookPath := filepath.Join(fullRepoPath, "hooks", "pre-receive")

	if info, err := os.Stat(preReceiveHookPath); err != nil || info.Mode()&0111 == 0 {
		return "", errors.New("pre-receive hook is missing or is not executable")
	}

	return fullRepoPath, nil
}
