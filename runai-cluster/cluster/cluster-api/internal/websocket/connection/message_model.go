package connection

type (
	MessageType   string
	MessageStatus string
	MessageKind   string
)

const (
	TypeTerminalSize MessageType = "terminal_size"

	KindRequest  MessageKind = "request"
	KindResponse MessageKind = "response"

	StatusSuccess MessageStatus = "Success"
	StatusFailure MessageStatus = "Failure"
)

type Message struct {
	Version    string         `json:"version"`
	Id         int            `json:"id,omitempty"`
	Message    string         `json:"message,omitempty"`
	Type       MessageType    `json:"type,omitempty"`
	Params     map[string]any `json:"params,omitempty"`
	Status     MessageStatus  `json:"status,omitempty"` // Status was used in the older cluster versions.
	StatusCode int            `json:"statusCode,omitempty"`
	Kind       MessageKind    `json:"kind,omitempty"`
}

func (r *Message) GetFloat64(key string, def float64) float64 {
	valueAny, ok := r.Params[key]
	if !ok {
		return def
	}

	value, ok := valueAny.(float64)
	if !ok {
		return def
	}

	return value
}

func (r *Message) GetInt(key string, def int) int {
	value := r.GetFloat64(key, float64(def))
	return int(value)
}
