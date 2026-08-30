package model

type InstallPhase int

const (
	NotInstalled InstallPhase = iota
	Installing
	Failed
	Installed
	Timeout
)
