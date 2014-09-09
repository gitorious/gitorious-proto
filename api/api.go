package api

type User struct {
	Username string
}

type RepositoryConfiguration struct {
	RealPath string

	SshCloneUrl  string
	HttpCloneUrl string
	GitCloneUrl  string

	CustomPreReceivePath  string
	CustomPostReceivePath string
	CustomUpdatePath      string
}

type PublicApi interface {
	GetUserInfo() (*User, error)
}

type InternalApi interface {
	GetRepositoryConfiguration(string) (*RepositoryConfiguration, error)
	AuthorizeRefPush(string, string, string, string) (bool, error)
}
