package auth

import "k8s.io/client-go/rest"

type Info struct {
	Cfg               *rest.Config
	Username          string
	Subject           string
	OriginalSubject   string
	Groups            []string
	Uid               *int64
	Gid               *int64
	CanSeeAllProjects bool
	PreferredUsername string
	RawToken          string
}
