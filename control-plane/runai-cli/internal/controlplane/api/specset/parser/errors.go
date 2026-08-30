package parser

import (
	"errors"
)

var (
	ErrParserNoDic       = errors.New("parser is missing dictionary")
	ErrParseValueIsEmpty = errors.New("value is an empty string")
)

const (
	ErrMsgDictionaryMustContainAllKeys = "dictionary must contain all keys"
	ErrMsgDictionaryDoesNotContainKey  = "dictionary does not contain key"
	ErrMsgMissingMandatoryFields       = "missing mandatory fields"
)
