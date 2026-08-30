package utils

import "errors"

var (
	ErrInvalidInput                  = errors.New("invalid input")
	ErrBannerTextRequiredWhenEnabled = errors.New("text is required when banner is enabled")
	ErrLogoTooLarge                  = errors.New("logo size is too big, limit is 128kb")
	ErrUnsupportedMediaType          = errors.New("unsupported file type, only PNG and SVG images are supported")
)

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}
