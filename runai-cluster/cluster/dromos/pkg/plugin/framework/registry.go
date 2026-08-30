package framework

// Registry stores all the plugins that are used by the framework.
type Registry map[string]Plugin

func NewRegistry(plugins ...Plugin) Registry {
	reg := Registry{}
	for _, plugin := range plugins {
		reg[plugin.Name()] = plugin
	}

	return reg
}
