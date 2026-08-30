package models

type Verbosity int

const (
	SlimVerbosity Verbosity = iota
	MetaDataVerbosity
	VerboseVerbosity
)
