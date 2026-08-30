package loader

type Chart struct{}

func Load(path string) (*Chart, error) { return &Chart{}, nil }
