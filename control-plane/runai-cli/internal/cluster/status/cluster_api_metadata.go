package status

type Features struct {
	MessageEnabled        bool `json:"message_enabled,omitempty"`
	TerminalResizeEnabled bool `json:"message_resize_enabled,omitempty"`
}

type Metadata struct {
	Version  string   `json:"version,omitempty"`
	Features Features `json:"features"`
}
