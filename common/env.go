package common

import (
	"fmt"
	"os"

	"gitorious.org/gitorious/gitorious-proto/api"
)

func CreateEnv(protocol, username string, repoConfig *api.RepoConfig) []string {
	env := os.Environ()

	// used by hooks
	env = append(env, "GITORIOUS_PROTO="+protocol)
	env = append(env, "GITORIOUS_USER="+username)
	env = append(env, fmt.Sprintf("GITORIOUS_REPOSITORY_ID=%v", repoConfig.RepositoryId))

	if repoConfig.SshCloneUrl != "" {
		env = append(env, "GITORIOUS_SSH_CLONE_URL="+repoConfig.SshCloneUrl)
	}

	if repoConfig.HttpCloneUrl != "" {
		env = append(env, "GITORIOUS_HTTP_CLONE_URL="+repoConfig.HttpCloneUrl)
	}

	if repoConfig.GitCloneUrl != "" {
		env = append(env, "GITORIOUS_GIT_CLONE_URL="+repoConfig.GitCloneUrl)
	}

	if repoConfig.CustomPreReceivePath != "" {
		env = append(env, "GITORIOUS_CUSTOM_PRE_RECEIVE_PATH="+repoConfig.CustomPreReceivePath)
	}

	if repoConfig.CustomPostReceivePath != "" {
		env = append(env, "GITORIOUS_CUSTOM_POST_RECEIVE_PATH="+repoConfig.CustomPostReceivePath)
	}

	if repoConfig.CustomUpdatePath != "" {
		env = append(env, "GITORIOUS_CUSTOM_UPDATE_PATH="+repoConfig.CustomUpdatePath)
	}

	return env
}

func Getenv(name, defaultValue string) string {
	value := os.Getenv(name)

	if value == "" {
		value = defaultValue
	}

	return value
}
