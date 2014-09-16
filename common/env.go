package common

import (
	"os"

	"gitorious.org/gitorious/gitorious-shell/api"
)

func CreateEnv(protocol, username string, repoConfig *api.RepoConfig) []string {
	env := os.Environ()

	env = append(env, "GITORIOUS_PROTO="+protocol)
	env = append(env, "GITORIOUS_USER="+username) // used by hooks

	// if repoConfig.CustomPreReceivePath != "" {
	// 	env = append(env, "GITORIOUS_CUSTOM_PRE_RECEIVE=1")
	// }

	return env
}

func Getenv(name, defaultValue string) string {
	value := os.Getenv(name)

	if value == "" {
		value = defaultValue
	}

	return value
}
