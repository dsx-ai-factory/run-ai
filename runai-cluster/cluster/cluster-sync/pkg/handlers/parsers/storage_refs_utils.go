package parsers

import (
	"context"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	field_types "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	cu "github.com/run-ai/runai/runai-cluster/common/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// storageRef is the internal name+kind reference both parse paths collect before
// resolving each one to its Kubernetes UID.
type storageRef struct {
	kind v1alpha1.StorageRefKind
	name string
}

// newObjectForKind returns an empty object of the given kind, or nil for an
// unrecognized kind so the caller can skip it instead of resolving against the
// wrong resource type.
func newObjectForKind(kind v1alpha1.StorageRefKind) client.Object {
	switch kind {
	case v1alpha1.StorageRefKindPVC:
		return &corev1.PersistentVolumeClaim{}
	case v1alpha1.StorageRefKindSecret:
		return &corev1.Secret{}
	case v1alpha1.StorageRefKindConfigMap:
		return &corev1.ConfigMap{}
	default:
		return nil
	}
}

// resolveStorageRefs resolves each name+kind reference to its live Kubernetes UID via the
// cached reader. References that are not found in cache yet are skipped so the workload
// still syncs without them; they get picked up on a later reconcile once the cache catches up.
func (p *WorkloadParser) resolveStorageRefs(ctx context.Context, namespace string, refs []storageRef) []openapi.ResourceRef {
	resolved := make([]openapi.ResourceRef, 0, len(refs))
	for _, ref := range refs {
		obj := newObjectForKind(ref.kind)
		if obj == nil {
			zap.S().Warnf("Skipping storage ref <%s/%s> in namespace <%s>: unsupported kind",
				ref.kind, ref.name, namespace)
			continue
		}
		if err := p.cachedReader.Get(ctx, client.ObjectKey{Namespace: namespace, Name: ref.name}, obj); err != nil {
			if client.IgnoreNotFound(err) != nil {
				zap.S().Warnf("Failed to resolve storage ref <%s/%s> in namespace <%s>, error: %s",
					ref.kind, ref.name, namespace, err.Error())
			}
			continue
		}
		resolved = append(resolved, openapi.ResourceRef{
			Uid:  ptr.To(string(obj.GetUID())),
			Name: ref.name,
			Kind: openapi.ResourceRefKind(ref.kind),
		})
	}
	return resolved
}

// nativeStorageRefs collects the existing-resource references from a native workload's
// map-backed StorageFields and container imagePullSecrets. PVC iteration mirrors the
// PvcBuilder so generated claim names line up; secret/configMap names come straight from
// each item value. Names that cannot be resolved later are dropped by resolveStorageRefs,
// so best-effort generated names are safe.
func nativeStorageRefs(workloadName, namespace string, sf field_groups.StorageFields, imagePullSecrets *field_types.StringField) []storageRef {
	var refs []storageRef

	pvcs := sf.Pvcs.Deref()
	idx := 0
	for _, nick := range cu.SortedKeys(pvcs) {
		pvc := pvcs[nick].Value
		if pvc == nil || ptr.Deref(pvc.Ephemeral, false) {
			continue
		}
		refs = appendStorageRef(refs, v1alpha1.StorageRefKindPVC, builders.GetPvcClaimName(pvc, workloadName, "", idx))
		idx++
	}

	secretVolumes := sf.SecretVolumes.Deref()
	for _, nick := range cu.SortedKeys(secretVolumes) {
		if item := secretVolumes[nick].Value; item != nil {
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindSecret, ptr.Deref(item.Name, ""))
		}
	}

	configMapVolumes := sf.ConfigMapVolumes.Deref()
	for _, nick := range cu.SortedKeys(configMapVolumes) {
		if item := configMapVolumes[nick].Value; item != nil {
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindConfigMap, ptr.Deref(item.Name, ""))
		}
	}

	gitSync := sf.GitSync.Deref()
	for _, nick := range cu.SortedKeys(gitSync) {
		if item := gitSync[nick].Value; item != nil {
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindSecret, ptr.Deref(item.SecretName, ""))
		}
	}

	s3buckets := sf.S3buckets.Deref()
	for _, nick := range cu.SortedKeys(s3buckets) {
		if item := s3buckets[nick].Value; item != nil {
			refs = appendStorageRef(refs, v1alpha1.StorageRefKindSecret, ptr.Deref(item.SecretName, ""))
		}
	}

	if imagePullSecrets != nil {
		names, err := builders.ParseImagePullSecrets(imagePullSecrets.Deref(),
			builders.WorkloadInfo{Name: workloadName, Namespace: namespace})
		if err != nil {
			zap.S().Warnf("Failed to parse imagePullSecrets for workload <%s/%s>, error: %s",
				namespace, workloadName, err.Error())
		} else {
			for _, name := range cu.SortedKeys(names) {
				refs = appendStorageRef(refs, v1alpha1.StorageRefKindSecret, name)
			}
		}
	}

	return dedupeStorageRefs(refs)
}

// appendStorageRef appends a name+kind reference, skipping empty names.
func appendStorageRef(refs []storageRef, kind v1alpha1.StorageRefKind, name string) []storageRef {
	if name == "" {
		return refs
	}
	return append(refs, storageRef{kind: kind, name: name})
}

func dedupeStorageRefs(refs []storageRef) []storageRef {
	seen := make(map[storageRef]bool, len(refs))
	deduped := make([]storageRef, 0, len(refs))
	for _, ref := range refs {
		if !seen[ref] {
			seen[ref] = true
			deduped = append(deduped, ref)
		}
	}
	return deduped
}
