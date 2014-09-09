package api

type Error int

const (
	UNAUTHENTICATED Error = iota + 1
)

func (e Error) Error() string {
	switch e {
	case UNAUTHENTICATED:
		return "invalid username or password"
	}

	return "unknown error"
}
