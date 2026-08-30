package airgapvalues

import "strings"

// values is a Helm values tree addressed by dotted paths.
type values map[string]any

// set writes value at a dotted path, creating the intermediate nodes.
func (v values) set(path string, value any) {
	keys := strings.Split(path, ".")
	node := v
	for _, key := range keys[:len(keys)-1] {
		child, ok := node[key].(values)
		if !ok {
			child = values{}
			node[key] = child
		}
		node = child
	}
	node[keys[len(keys)-1]] = value
}
