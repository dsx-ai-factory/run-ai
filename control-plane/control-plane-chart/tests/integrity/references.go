package integrity

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reference is one (secret, key) pair consumed by a control-plane service.
// Key is "*" for envFrom or full-secret volume mounts where individual keys
// are not listed; only secret existence can be validated for those.
type Reference struct {
	Secret    string
	Key       string
	Service   string // e.g., "Deployment/runai-backend-cluster-service"
	Container string
	Field     string // e.g., "env[FOO]", "volume[creds].items[token]"
}

const chartLabelPrefix = "control-plane-"

func collectReferences(m *Manifests) []Reference {
	var refs []Reference
	walk := func(meta metav1.ObjectMeta, kind string, spec corev1.PodSpec) {
		if !strings.HasPrefix(meta.Labels["helm.sh/chart"], chartLabelPrefix) {
			return
		}
		svc := fmt.Sprintf("%s/%s", kind, meta.Name)
		refs = append(refs, walkPodSpec(svc, spec)...)
	}
	for _, d := range m.Deployments {
		walk(d.ObjectMeta, "Deployment", d.Spec.Template.Spec)
	}
	for _, s := range m.StatefulSets {
		walk(s.ObjectMeta, "StatefulSet", s.Spec.Template.Spec)
	}
	for _, d := range m.DaemonSets {
		walk(d.ObjectMeta, "DaemonSet", d.Spec.Template.Spec)
	}
	for _, j := range m.Jobs {
		walk(j.ObjectMeta, "Job", j.Spec.Template.Spec)
	}
	for _, c := range m.CronJobs {
		walk(c.ObjectMeta, "CronJob", c.Spec.JobTemplate.Spec.Template.Spec)
	}
	return refs
}

func walkPodSpec(svc string, spec corev1.PodSpec) []Reference {
	var refs []Reference
	for _, c := range spec.InitContainers {
		refs = append(refs, walkContainer(svc, c)...)
	}
	for _, c := range spec.Containers {
		refs = append(refs, walkContainer(svc, c)...)
	}
	for _, v := range spec.Volumes {
		refs = append(refs, walkVolume(svc, v)...)
	}
	return refs
}

func walkContainer(svc string, c corev1.Container) []Reference {
	var refs []Reference
	for _, e := range c.Env {
		if e.ValueFrom == nil || e.ValueFrom.SecretKeyRef == nil {
			continue
		}
		skr := e.ValueFrom.SecretKeyRef
		if skr.Optional != nil && *skr.Optional {
			continue
		}
		refs = append(refs, Reference{
			Secret: skr.Name, Key: skr.Key,
			Service: svc, Container: c.Name,
			Field: fmt.Sprintf("env[%s]", e.Name),
		})
	}
	for i, ef := range c.EnvFrom {
		if ef.SecretRef == nil {
			continue
		}
		if ef.SecretRef.Optional != nil && *ef.SecretRef.Optional {
			continue
		}
		refs = append(refs, Reference{
			Secret: ef.SecretRef.Name, Key: "*",
			Service: svc, Container: c.Name,
			Field: fmt.Sprintf("envFrom[%d]", i),
		})
	}
	return refs
}

func walkVolume(svc string, v corev1.Volume) []Reference {
	var refs []Reference
	if v.Secret != nil {
		refs = append(refs, secretVolumeRefs(svc, v.Name, v.Secret.SecretName, v.Secret.Items, v.Secret.Optional)...)
	}
	if v.Projected != nil {
		for i, src := range v.Projected.Sources {
			if src.Secret == nil {
				continue
			}
			field := fmt.Sprintf("volume[%s].projected[%d]", v.Name, i)
			refs = append(refs, projectedSecretRefs(svc, field, src.Secret)...)
		}
	}
	return refs
}

func secretVolumeRefs(svc, volume, name string, items []corev1.KeyToPath, optional *bool) []Reference {
	if optional != nil && *optional {
		return nil
	}
	if len(items) == 0 {
		return []Reference{{
			Secret: name, Key: "*",
			Service: svc,
			Field:   fmt.Sprintf("volume[%s]", volume),
		}}
	}
	var refs []Reference
	for _, it := range items {
		refs = append(refs, Reference{
			Secret: name, Key: it.Key,
			Service: svc,
			Field:   fmt.Sprintf("volume[%s].items[%s]", volume, it.Key),
		})
	}
	return refs
}

func projectedSecretRefs(svc, field string, src *corev1.SecretProjection) []Reference {
	if src.Optional != nil && *src.Optional {
		return nil
	}
	if len(src.Items) == 0 {
		return []Reference{{
			Secret: src.Name, Key: "*",
			Service: svc, Field: field,
		}}
	}
	var refs []Reference
	for _, it := range src.Items {
		refs = append(refs, Reference{
			Secret: src.Name, Key: it.Key,
			Service: svc,
			Field:   fmt.Sprintf("%s.items[%s]", field, it.Key),
		})
	}
	return refs
}
