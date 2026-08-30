package collector

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sResourceKind string

const (
	DaemonSet   K8sResourceKind = "DaemonSet"
	Deployment  K8sResourceKind = "Deployment"
	ReplicaSet  K8sResourceKind = "ReplicaSet"
	StatefulSet K8sResourceKind = "StatefulSet"
	Job         K8sResourceKind = "Job"
	CronJob     K8sResourceKind = "CronJob"
	Pod         K8sResourceKind = "Pod"
)

var (
	// Regex patterns to identify the origin of pods created by controllers
	// In CronJob, the suffix is an 8-digit epoch timestamp in minutes
	cronJobRegex = regexp.MustCompile(`^(.+)-[0-9]{8}$`)

	resourceKindToFolderNamingMap = map[K8sResourceKind]string{
		DaemonSet:   "daemon-set",
		ReplicaSet:  "replica-set",
		Deployment:  "deployment",
		StatefulSet: "stateful-set",
		Job:         "job",
		CronJob:     "cron-job",
	}
)

const (
	podFileNameWithPrefixFormat = "%s%s_%s.log"
	podFileNameFormat           = "%s%s.log"
)

type PodInfo struct {
	corev1.Pod
}

func NewPodInfo(pod *corev1.Pod) *PodInfo {
	return &PodInfo{Pod: *pod}
}

func (p *PodInfo) getLogFilePathInArchive(containerName string, previous bool) string {
	base := filepath.Join(p.Namespace, p.Name)
	kind, ownerName := p.getPodOwnerKindAndName()
	if kind != Pod {
		base = filepath.Join(p.Namespace, fmt.Sprintf("%s_%s", resourceKindToFolderNamingMap[kind], ownerName), p.Name)
	}

	logFile := p.getPodLogFileName(containerName, kind == DaemonSet, previous)
	return filepath.Join(base, logFile)
}

func (p *PodInfo) getPodLogFileName(containerName string, isDaemonSet bool, previous bool) string {
	prefix := ""
	if previous {
		prefix = "previous_"
	}

	if isDaemonSet {
		return fmt.Sprintf(podFileNameWithPrefixFormat, prefix, p.Spec.NodeName, containerName)
	}

	return fmt.Sprintf(podFileNameFormat, prefix, containerName)
}

func (p *PodInfo) getPodOwnerKindAndName() (K8sResourceKind, string) {
	if len(p.OwnerReferences) == 0 {
		return Pod, p.Name
	}

	owner := p.OwnerReferences[0]
	ownerKind := K8sResourceKind(owner.Kind)
	switch ownerKind {
	case Job:
		return guessJobOrigin(&owner)
	case ReplicaSet:
		return guessReplicaSetOrigin(&owner)
	default:
		return ownerKind, owner.Name
	}
}

func guessJobOrigin(owner *metav1.OwnerReference) (K8sResourceKind, string) {
	if matches := cronJobRegex.FindStringSubmatch(owner.Name); len(matches) == 2 {
		return CronJob, matches[1]
	}
	return K8sResourceKind(owner.Kind), owner.Name
}

func guessReplicaSetOrigin(owner *metav1.OwnerReference) (K8sResourceKind, string) {
	name := owner.Name

	lastDashIndex := strings.LastIndex(name, "-")
	if lastDashIndex == -1 {
		return ReplicaSet, name
	}

	prefix := name[:lastDashIndex]
	suffix := name[lastDashIndex+1:]

	if isValidReplicaSetHash(suffix) {
		return Deployment, prefix
	}
	return ReplicaSet, name
}

// isValidReplicaSetHash verifies that the suffix length is between 7 and 10 characters,
// and contains at least one letter and one digit. It only allows lowercase letters and digits.
func isValidReplicaSetHash(suffix string) bool {
	if len(suffix) < 7 || len(suffix) > 10 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for i := 0; i < len(suffix); i++ {
		if hasDigit && hasLetter {
			return true
		}

		c := suffix[i]
		switch {
		case c >= 'a' && c <= 'z':
			hasLetter = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			return false
		}
	}

	return hasLetter && hasDigit
}
