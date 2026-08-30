package actions

import "github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"

// SystemNodeAffinityValues renders the helm --set lines for a node affinity that
// targets label (Exists) under the given values prefix (e.g. "global.affinity"
// or "runai-operator.config.global.affinity"). When required is true it emits a
// hard requiredDuringScheduling term; otherwise a soft preferred term.
func SystemNodeAffinityValues(prefix, label string, required bool) []string {
	base := prefix + ".nodeAffinity"
	if required {
		term := base + ".requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0]"
		return []string{
			term + ".key=" + label,
			term + ".operator=Exists",
		}
	}
	pref := base + ".preferredDuringSchedulingIgnoredDuringExecution[0]"
	return []string{
		pref + ".weight=1",
		pref + ".preference.matchExpressions[0].key=" + label,
		pref + ".preference.matchExpressions[0].operator=Exists",
	}
}

// selectSystemNodeNames returns the names of nodes matching the system-node
// selection: when names is non-empty, the named nodes; otherwise the nodes whose
// labels match selector. Control-plane nodes are always excluded.
func selectSystemNodeNames(nodes []orchestrator2.NodeInfo, names []string, selector map[string]string) []string {
	wanted := make(map[string]bool, len(names))
	for _, n := range names {
		wanted[n] = true
	}
	var out []string
	for _, n := range nodes {
		if isControlPlaneNode(n.Roles) {
			continue
		}
		if len(names) > 0 {
			if wanted[n.Name] {
				out = append(out, n.Name)
			}
			continue
		}
		if labelsMatch(n.Labels, selector) {
			out = append(out, n.Name)
		}
	}
	return out
}

// scopeToSystemNodes keeps only the nodes matching the configured system-node
// selection — the same set that receives the system-node affinity/label.
func scopeToSystemNodes(nodes []orchestrator2.NodeInfo, names []string, selector map[string]string) []orchestrator2.NodeInfo {
	wanted := make(map[string]bool)
	for _, name := range selectSystemNodeNames(nodes, names, selector) {
		wanted[name] = true
	}
	scoped := make([]orchestrator2.NodeInfo, 0, len(wanted))
	for _, n := range nodes {
		if wanted[n.Name] {
			scoped = append(scoped, n)
		}
	}
	return scoped
}

// missingNamedNodes returns the requested node names that do not correspond to
// any node in the cluster (likely typos). It is empty for selector-based
// selection (names is empty).
func missingNamedNodes(nodes []orchestrator2.NodeInfo, names []string) []string {
	exists := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		exists[n.Name] = true
	}
	var missing []string
	for _, name := range names {
		if !exists[name] {
			missing = append(missing, name)
		}
	}
	return missing
}

func isControlPlaneNode(roles []string) bool {
	for _, r := range roles {
		if r == "control-plane" || r == "master" {
			return true
		}
	}
	return false
}

func labelsMatch(labels, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
