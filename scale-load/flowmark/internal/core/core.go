package core

import "context"

type Scenario interface {
	Name() string
	Run(ctx context.Context, env Env, cfg Config) (Section, error)
}

type Env struct {
	KubeconfigPath  string
	ControlPlaneURL string
	Username        string
	Password        string
	Token           string
}

type Config struct {
	Suite string
	Knobs map[string]any
}

type Category string

const (
	CategoryScale Category = "scale"
	CategoryLoad  Category = "load"
)

type Section struct {
	Category Category
	Metrics  map[string]any
	Failed   bool
	Error    string
}
