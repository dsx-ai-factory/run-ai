package spec

import (
	"fmt"
	"math"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction/info"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/utils/ptr"
)

const (
	GroupNameJob = "Job Type"

	AttributeNameInteractive = "interactive"

	AttributeUsageInteractive = "Mark this job as interactive"

	ColonString = ":"

	VolumeHostPath         = 2
	VolumeHostPathReadFlag = 3
	VolumeNfs              = 4

	HostPathReadOnly  = "path=%s,mount=%s"
	HostPathReadWrite = "path=%s,mount=%s,readwrite"
	NfsReadWrite      = "path=%s,server=%s,mountpath=%s,readwrite"
	NfsReadOnly       = "path=%s,server=%s,mountpath=%s"

	ReadOnlyFlag  = "ro"
	ReadWriteFlag = "rw"

	TargetAttribute     = "target="
	PathAttribute       = "path="
	NameAttribute       = "name="
	SourceAttribute     = "source="
	RepositoryAttribute = "repository="
	BranchAttribute     = "branch="
	RevAttribute        = "rev="

	RandomSuffixLength = 10
)

type SubmitSpec struct {
	GeneralSpec
	Interactive bool
	name        string
}

func NewSubmitSpec() *SubmitSpec {
	return &SubmitSpec{
		GeneralSpec: GeneralSpec{workloadType: spectype.Submit},
		Interactive: false,
	}
}

func (s *SubmitSpec) SetType(workloadType spectype.Workload) {
	s.workloadType = workloadType
}

func (s *SubmitSpec) SetFlagExtended(cmd *cobra.Command, project *string) {
	s.getGrouper(cmd)

	// backward compatibility flags
	job := s.grouper.GetOrAddFlagSet(GroupNameJob)
	jobFlagSet(job, &s.Interactive)

	s.SetFlags(cmd, project)
}

func (s *SubmitSpec) BackwardCompatabilityMutate(name string, oldStylePvc []string, bcServiceType []string, bcCustomUrl string) error {
	if name != "" {
		s.name = name
	} else {
		s.name = string(s.GetType())
	}

	err := s.matchGpuValueToField()
	if err != nil {
		return err
	}
	err = s.mutateVolumeToHostPath()
	if err != nil {
		return err
	}
	s.mutateS3()
	s.mutateToleration()
	s.gitMutate()
	s.handleBackwardCompatiblePVC(oldStylePvc)
	err = s.MutateServiceType(bcServiceType, bcCustomUrl)
	if err != nil {
		return err
	}
	return nil
}
func (s *SubmitSpec) handleBackwardCompatiblePVC(oldStylePvc []string) {
	// simplest way to do this, but we need to keep the old style pvc for the backward compatibility
	// remove this as soon as possible
	s.flags.Storage.BCStylePvc = oldStylePvc
}

func (s *SubmitSpec) MutateServiceType(serviceType []string, customUrlArg string) error {
	if customUrlArg != "" && len(serviceType) == 0 {
		// custom url provided without service type
		return fmt.Errorf("custom-url provided without a service-type")
	}
	if len(serviceType) == 0 {
		// not service type provided meaning no backward compatibility needed
		return nil
	}
	bcNetworkSubmitArgs, err := ParseServiceTypeParams(serviceType, s.flags.Network.Ports, customUrlArg)
	if err != nil {
		return err
	}
	// after parsing the old style service type, we need to remove the old style ports
	s.flags.Network.Ports = nil
	if len(bcNetworkSubmitArgs.nodePorts) != 0 {
		for _, portStr := range bcNetworkSubmitArgs.nodePorts {
			newPortString, parseErr := mutateNodePort(portStr)
			if parseErr != nil {
				return parseErr
			}
			s.flags.Network.Ports = append(s.flags.Network.Ports, newPortString)
		}
	}
	if len(bcNetworkSubmitArgs.externalUrlArgs) != 0 {
		for _, externalUrlArg := range bcNetworkSubmitArgs.externalUrlArgs {
			externalUrlString, err2 := mutateExternalUrl(externalUrlArg)
			if err2 != nil {
				return err2
			}
			s.flags.Network.ExternalUrls = append(s.flags.Network.ExternalUrls, externalUrlString)
		}
	}
	return nil

}

func mutateExternalUrl(externalUrlArg ExternalUrlArg) (string, error) {
	_, containerPort, parseErr := parseServicePortLiterals(externalUrlArg.port, false)
	if parseErr != nil {
		return "", parseErr
	}
	externalUrlString := fmt.Sprintf("container=%d", containerPort)
	if externalUrlArg.customUrl != "" {
		externalUrlString += fmt.Sprintf(",url=%s", externalUrlArg.customUrl)
	}
	return externalUrlString, nil
}

func mutateNodePort(portStr string) (string, error) {
	externalPort, containerPort, err := parseServicePortLiterals(portStr, true)
	if err != nil {
		return "", err
	}
	newPortString := fmt.Sprintf("service-type=NodePort,container=%d,external=%d", containerPort, externalPort)
	return newPortString, nil
}

func (s *SubmitSpec) matchGpuValueToField() error {
	gpu := s.flags.Compute.GpuPortionRequest
	if !gpu.IsSet() {
		return nil
	}
	gpuValue := gpu.Get()
	if gpuValue != nil && *gpuValue > 0 {
		if *gpuValue <= 1 {
			// gpu is portion
			return nil
		}
		if *gpuValue != math.Trunc(*gpuValue) {
			return fmt.Errorf("gpu value cannot be a fraction in whole numbers: %.2f", *gpuValue)
		}
		gpuInt32 := int32(*gpuValue)
		s.flags.Compute.GpuDeviceRequest.Set(ptr.To(gpuInt32))
		s.flags.Compute.GpuPortionRequest.Unset()
	}
	return nil
}

func (s *SubmitSpec) mutateVolumeToHostPath() error {
	var (
		volume       string
		hostPathList []string
		nfsList      []string
	)
	storage := &s.flags.Storage
	volumeList := &storage.Volume
	if len(*volumeList) == 0 {
		return nil
	}

	for _, v := range *volumeList {
		vSplit := strings.Split(v, ColonString)
		vSplintLen := len(vSplit)
		switch vSplintLen {
		case VolumeHostPath:
			volume = fmt.Sprintf(HostPathReadOnly, vSplit[0], vSplit[1])
			hostPathList = append(hostPathList, volume)
		case VolumeHostPathReadFlag:
			rwFlag, err := checkForReadWriteFlag(vSplit[2])
			if err != nil {
				return err
			}
			if rwFlag {
				volume = fmt.Sprintf(HostPathReadWrite, vSplit[0], vSplit[1])
				hostPathList = append(hostPathList, volume)
			} else {
				volume = fmt.Sprintf(HostPathReadOnly, vSplit[0], vSplit[1])
				hostPathList = append(hostPathList, volume)
			}
		case VolumeNfs:
			rwFlag, err := checkForReadWriteFlag(vSplit[2])
			if err != nil {
				return err
			}
			if rwFlag {
				volume = fmt.Sprintf(NfsReadWrite, vSplit[0], vSplit[3], vSplit[1])
				nfsList = append(nfsList, volume)
			} else {
				volume = fmt.Sprintf(NfsReadOnly, vSplit[0], vSplit[3], vSplit[1])
				nfsList = append(nfsList, volume)
			}
		}
	}
	storage.Nfs = nfsList
	storage.HostPath = hostPathList
	return nil
}

func (s *SubmitSpec) mutateS3() {
	s3List := &s.flags.Storage.S3
	if len(*s3List) == 0 {
		return
	}
	for i := range *s3List {
		replaceAttribute(s3List, i, TargetAttribute, PathAttribute)
	}
}

func (s *SubmitSpec) mutateToleration() {
	tList := &s.flags.Scheduling.Toleration
	for i := range *tList {
		if !strings.Contains((*tList)[i], NameAttribute) {
			(*tList)[i] += fmt.Sprintf(",%s%s-%s", NameAttribute, s.name, info.GenerateRandomString(RandomSuffixLength))
		}
	}
}

func (s *SubmitSpec) gitMutate() {
	gitList := &s.flags.Storage.GitSync
	for i := range *gitList {
		replaceAttribute(gitList, i, SourceAttribute, RepositoryAttribute)
		replaceAttribute(gitList, i, TargetAttribute, PathAttribute)
		replaceAttribute(gitList, i, BranchAttribute, RevAttribute)
	}
}

func replaceAttribute(list *[]string, index int, lookup, replace string) {
	if strings.Contains((*list)[index], lookup) {
		(*list)[index] = strings.ReplaceAll((*list)[index], lookup, replace)
	}
}

func checkForReadWriteFlag(rwFlag string) (bool, error) {
	switch rwFlag {
	case ReadOnlyFlag:
		return false, nil
	case specset.EmptyString, ReadWriteFlag:
		return true, nil
	}
	return false, fmt.Errorf("could not recognise readwrite flag (ro|rw|empty), flag: %s", rwFlag)
}

func jobFlagSet(job *pflag.FlagSet, interactive *bool) {
	job.BoolVar(interactive, AttributeNameInteractive, false, AttributeUsageInteractive)
}
