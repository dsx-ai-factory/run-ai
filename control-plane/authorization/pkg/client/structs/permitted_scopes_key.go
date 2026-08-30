package structs

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	PermittedScopesKey = contextKey("permitted-scopes")
)
