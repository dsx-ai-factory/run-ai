package error

type ErrorNotFound struct {
	Message string `json:"message,omitempty"`
}

func (e ErrorNotFound) Error() string {
	return e.Message
}

type ErrorBadRequest struct {
	Message string `json:"message,omitempty"`
}

func (e ErrorBadRequest) Error() string {
	return e.Message
}

type ErrorInternal struct {
	Message string `json:"message,omitempty"`
}

func (e ErrorInternal) Error() string {
	return e.Message
}

type ErrorConflict struct {
	Message string `json:"message,omitempty"`
}

func (e ErrorConflict) Error() string {
	return e.Message
}
