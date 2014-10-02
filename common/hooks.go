package common

import (
	"os"
	"path/filepath"
)

func PreReceiveHookExists(fullRepoPath string) bool {
	preReceiveHookPath := filepath.Join(fullRepoPath, "hooks", "pre-receive")

	if info, err := os.Stat(preReceiveHookPath); err != nil || info.Mode()&0111 == 0 {
		return false
	}

	return true
}
