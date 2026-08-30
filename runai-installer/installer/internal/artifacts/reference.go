package artifacts

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
)

// MirrorRef returns where an image is pushed in the destination registry.
func MirrorRef(destination string, entry artifactlock.ImageEntry) (ref, tag string, err error) {
	tag, err = SourceTag(entry.Image.Source)
	if err != nil {
		return "", "", err
	}
	return RegistryPrefix(destination, entry.Group) + "/" + entry.Image.Name + ":" + tag, tag, nil
}

func RegistryPrefix(destination, group string) string {
	return strings.TrimSuffix(destination, "/") + "/" + group
}

// SourceTag returns the tag a reference pins, ignoring both a registry port and
// a trailing digest.
func SourceTag(source string) (string, error) {
	name := source[strings.LastIndex(source, "/")+1:]
	if i := strings.Index(name, "@"); i >= 0 {
		name = name[:i]
	}
	i := strings.LastIndex(name, ":")
	if i < 0 || i == len(name)-1 {
		return "", fmt.Errorf("source %q carries no tag to mirror", source)
	}
	return name[i+1:], nil
}
