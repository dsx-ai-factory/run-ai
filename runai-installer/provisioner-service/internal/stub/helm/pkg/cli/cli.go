package cli

type EnvSettings struct {
	namespace string
}

func New() *EnvSettings {
	return &EnvSettings{namespace: "default"}
}

func (e *EnvSettings) Namespace() string { return e.namespace }
func (e *EnvSettings) SetNamespace(ns string) {
	e.namespace = ns
}

type RESTGetter struct{}

func (e *EnvSettings) RESTClientGetter() *RESTGetter { return &RESTGetter{} }
