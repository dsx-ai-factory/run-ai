package action

type Configuration struct{}

func NewInstall(cfg *Configuration) *Install {
	return &Install{}
}

type Install struct {
	Namespace   string
	ReleaseName string
}
