package protocol

type ProtocolHandler interface {
	GetUsername() (string, error)
	ParseCommand() (string, string, error)
	RunProxy(RepositoryConfiguration, string, string) (string, error)
}
