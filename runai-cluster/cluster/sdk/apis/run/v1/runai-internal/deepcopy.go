package internal

import "encoding/json"

// This file is meant to define DeepCopyInto and DeepCopy functions for external types.
func (in *ClusterControlPlaneConfig) DeepCopyInto(out *ClusterControlPlaneConfig) {
	if in == nil || out == nil {
		return
	}
	b, _ := json.Marshal(in)
	json.Unmarshal(b, out)
}

func (in *ClusterControlPlaneConfig) DeepCopy() *ClusterControlPlaneConfig {
	if in == nil {
		return nil
	}
	out := new(ClusterControlPlaneConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *ClusterFullConfig) DeepCopyInto(out *ClusterFullConfig) {
	if in == nil || out == nil {
		return
	}
	b, _ := json.Marshal(in)
	json.Unmarshal(b, out)
}

func (in *ClusterFullConfig) DeepCopy() *ClusterFullConfig {
	if in == nil {
		return nil
	}
	out := new(ClusterFullConfig)
	in.DeepCopyInto(out)
	return out
}
