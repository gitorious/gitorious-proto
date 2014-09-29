package common

import (
	"os"
	"strings"
	"testing"

	"gitorious.org/gitorious/gitorious-proto/api"
)

func assertPresence(env []string, expected string, t *testing.T) {
	for _, envVar := range env {
		if envVar == expected {
			return
		}
	}

	t.Errorf("expected %v in env but not found", expected)
}

func assertAbsence(env []string, notExpected string, t *testing.T) {
	for _, envVar := range env {
		pair := strings.Split(envVar, "=")
		if pair[0] == notExpected {
			t.Errorf("not expected %v in env but found", notExpected)
			return
		}
	}
}

func TestCreateEnv(t *testing.T) {
	repoConfig := &api.RepoConfig{RepositoryId: 123}

	env := CreateEnv("ssh", "sickill", repoConfig)

	// make sure it is based on the existing environment
	assertPresence(env, "HOME="+os.Getenv("HOME"), t)

	// ensure required vars are set
	assertPresence(env, "GITORIOUS_PROTO=ssh", t)
	assertPresence(env, "GITORIOUS_USER=sickill", t)
	assertPresence(env, "GITORIOUS_REPOSITORY_ID=123", t)

	// ensure optional vars are not set
	assertAbsence(env, "GITORIOUS_SSH_CLONE_URL", t)
	assertAbsence(env, "GITORIOUS_HTTP_CLONE_URL", t)
	assertAbsence(env, "GITORIOUS_GIT_CLONE_URL", t)
	assertAbsence(env, "GITORIOUS_CUSTOM_PRE_RECEIVE_PATH", t)
	assertAbsence(env, "GITORIOUS_CUSTOM_POST_RECEIVE_PATH", t)
	assertAbsence(env, "GITORIOUS_CUSTOM_UPDATE_PATH", t)

	repoConfig = &api.RepoConfig{
		RepositoryId: 123,

		SshCloneUrl:  "ssh-clone-url",
		HttpCloneUrl: "http-clone-url",
		GitCloneUrl:  "git-clone-url",

		CustomPreReceivePath:  "custom-pre-receive",
		CustomPostReceivePath: "custom-post-receive",
		CustomUpdatePath:      "custom-update",
	}

	env = CreateEnv("ssh", "sickill", repoConfig)

	// ensure optional vars are not set
	assertPresence(env, "GITORIOUS_SSH_CLONE_URL=ssh-clone-url", t)
	assertPresence(env, "GITORIOUS_HTTP_CLONE_URL=http-clone-url", t)
	assertPresence(env, "GITORIOUS_GIT_CLONE_URL=git-clone-url", t)
	assertPresence(env, "GITORIOUS_CUSTOM_PRE_RECEIVE_PATH=custom-pre-receive", t)
	assertPresence(env, "GITORIOUS_CUSTOM_POST_RECEIVE_PATH=custom-post-receive", t)
	assertPresence(env, "GITORIOUS_CUSTOM_UPDATE_PATH=custom-update", t)
}
