package thirdparty

type Image struct {
	Name       string
	Repository string
	Tag        string
}

// TODO: This is a temporary solution, until we normalize the 3rd part images list.
// A ticket has been opened and will be worked on next.
var images = map[string][]Image{
	"runai-control-plane": {
		{Name: "keycloak", Repository: "quay.io/keycloak/keycloak", Tag: "26.6"},
		{Name: "nats", Repository: "docker.io/bitnamilegacy/nats", Tag: "2.11.8-debian-12-r0"},
		{Name: "postgres", Repository: "docker.io/postgres", Tag: "16-alpine"},
		{Name: "postgresql", Repository: "docker.io/bitnamilegacy/postgresql", Tag: "16"},
		{Name: "thanos", Repository: "docker.io/bitnamilegacy/thanos", Tag: "0.38.0-debian-12-r3"},
		{Name: "traefik", Repository: "docker.io/traefik", Tag: "v3.7"},
		{Name: "wait-ubi-minimal", Repository: "registry.access.redhat.com/ubi9/ubi-minimal", Tag: "9.8"},
	},
	"runai-cluster": {
		{Name: "git-sync", Repository: "registry.k8s.io/git-sync/git-sync", Tag: "v4.7.0"},
		{Name: "valkey", Repository: "docker.io/valkey/valkey", Tag: "9.0.4"},
		{Name: "wait-ubi-minimal", Repository: "registry.access.redhat.com/ubi9/ubi-minimal", Tag: "9.8"},
	},
}

func ForLock(name string) []Image {
	return append([]Image(nil), images[name]...)
}

func (image Image) Source() string {
	return image.Repository + ":" + image.Tag
}
