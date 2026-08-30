package pvc

import (
	"fmt"
	"regexp"

	"k8s.io/utils/ptr"

	"strings"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// delete this file after deprecated the old pvc format

// dns1123LabelFmt defines the format for DNS-1123 storage class names.
// The pattern enforces:
//   - Must start with: [a-z0-9] (lowercase letter or number)
//   - Optional middle: [-a-z0-9]* (any number of hyphens, lowercase letters, or numbers)
//   - If middle exists, must end with: [a-z0-9] (lowercase letter or number)
//
// Valid examples:
//   - "my-storage-class"
//   - "storage1"
//   - "a-b-c"
//   - "a"           // single character
//   - "123-abc"
//   - "storage-123-class"
//
// Invalid examples:
//   - "My-Storage-Class"  // uppercase not allowed
//   - "storage_class"     // underscores not allowed
//   - "-storage-class"    // cannot start with hyphen
//   - "storage-class-"    // cannot end with hyphen
//   - "storage.class"     // dots not allowed
//   - "Storage.Class"     // uppercase and dots not allowed
const (
	dns1123LabelFmt          = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	AccessModeReadOnlyParam  = "ro"
	AccessModeReadWriteParam = "rw"

	TwoParams   = 2
	ThreeParams = 3
	FourParams  = 4
)

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// resourceRequestRegex validates Kubernetes resource quantity strings.
// The pattern has two parts:
//
//  1. ([+-]?[0-9.]+)           // The number part
//     - Optional sign: [+-]?
//     - Numbers with optional decimal: [0-9.]+
//
//  2. ([eEinumkKMGTP]*[-+]?[0-9]*) // The unit/exponent part
//     - Optional unit suffixes: [eEinumkKMGTP]*
//     - Optional scientific notation: [-+]?[0-9]*
//
// Valid examples:
//   - "5Gi"      // 5 gibibytes
//   - "1.5Gi"    // 1.5 gibibytes
//   - "500Mi"    // 500 mebibytes
//   - "2.5G"     // 2.5 gigabytes
//   - "100M"     // 100 megabytes
//   - "1000Ki"   // 1000 kibibytes
//   - "1e6"      // scientific notation
//   - "1.5e3"    // decimal with scientific notation
//   - "+1.5Gi"   // explicit positive sign
//   - "1Ti"      // 1 tebibyte
//   - "1Pi"      // 1 pebibyte
//
// Invalid examples:
//   - "5g"       // lowercase 'g' not allowed
//   - "1.5.5Gi"  // multiple decimal points
//   - "Gi"       // must start with number
//   - "5GiB"     // invalid suffix
//   - "5 Gi"     // spaces not allowed
//   - "5GB"      // invalid suffix
//   - "1."       // decimal point must be followed by numbers
//   - ".5"       // must start with number
//   - "5Ki2"     // numbers after unit not allowed
var resourceRequestRegex = regexp.MustCompile("^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$")

func ListDeprecatedOldStyle(pvcsList []string) ([]openapi.PvcInstance, error) {
	pvcs := make([]openapi.PvcInstance, 0)
	for _, pvcStr := range pvcsList {
		pvc, err := NewFromDeprecatedOldStyle(pvcStr)
		if err != nil {
			return nil, err
		}
		if pvc != nil {
			pvcs = append(pvcs, *pvc)
		}
	}
	return pvcs, nil
}

// handle old backwords compatibility pvc parser
// this should be delete one we deprecated the old pvc format
func NewFromDeprecatedOldStyle(pvcStr string) (*openapi.PvcInstance, error) {
	pvcParams := strings.Split(pvcStr, ":")
	if len(pvcParams) < 2 || len(pvcParams) > 4 {
		return nil, fmt.Errorf("the --pvc parameter can be given in either of the forms: <storage_class>[optional]:<size>:<container_mount_path>:<access_mode>[optional] " +
			"or <existing_pvc_name>:<container_mount_path>:<access_mode>[optional] \n If any field is left blank the delimiting ':' must still be passed")
	}

	handleAsExistingPvc := extrapolateParamType(pvcParams)
	if handleAsExistingPvc {
		return handleExistingPvc(pvcParams)
	} else {
		return handleDynamicProvisionPvc(pvcParams)
	}
}

// Input: (surrounded by [] means optional)
//
//	[0]          [1]           [2]                [3]
//
// --pvc [<storage_class>]:<size>:<container_mount_path>:[<access_mode>]
// Or
//
//	[0]					[1]					[2]
//
// --pvc <existing_pvc_name>:<container_mount_path>:[<access_mode>]
func extrapolateParamType(pvcParams []string) (handleAsExistingPvc bool) {
	switch len(pvcParams) {
	case TwoParams:
		// len == 2 can either be 'existing-pvc:/path' or error
		handleAsExistingPvc = true
	case FourParams:
		// len == 4 can either be 'storageClass:2Gi:/path:ro', ':2Gi:/path:ro' or error
		handleAsExistingPvc = false
	default:
		// len == 3 is the tricky part since both parameter forms can have length 3. for instance ':2Gi:/path' and 'my-pvc:/path:ro'
		// In this case 2 checks can be made to tell us which is which:
		if pvcParams[0] == "" {
			// If params[0] == "" then we treat as dynamic provision form since only there we allow the first part to be empty (in order to use the default storage class)
			handleAsExistingPvc = false
		} else if validationErr := ValidateStorageResourceRequest(pvcParams[1]); validationErr == nil {
			// If params[1] is formatted like a resource request then we have to treat as dynamic provision
			handleAsExistingPvc = false
		} else if pvcParams[2] == AccessModeReadOnlyParam || pvcParams[2] == AccessModeReadWriteParam {
			// If params[2] == ro || params[2] == rw   we';; treat as existing pvc (since the params are of form '?:?:rw/ro'
			handleAsExistingPvc = true
		} else {
			// Otherwise, we have a param string of the form '?:?:?' and we know:
			// 		[0] != "",
			//		[1] is not a resource request,
			//		[2] != ro/rw
			//
			// So it can be either:
			//   - an erroneous dynamic-pvc param, like 'storage-class:badly-formatted-2g:/path'
			//   - an existing-pvc param, like 'pvc-name:/path:'
			//
			// The safest bet is to try it like an existing pvc.
			handleAsExistingPvc = true
		}
	}
	return handleAsExistingPvc
}

// Assumption: len(pvcParams) == 2 || len(pvcParams) == 3
func handleExistingPvc(pvcParams []string) (*openapi.PvcInstance, error) {
	var err error
	if len(pvcParams) == TwoParams {
		// Always normalize to len() == 3 to make it easier for us in the helm chart
		pvcParams = append(pvcParams, "")
	}
	claimName, path, access := pvcParams[0], pvcParams[1], pvcParams[2]
	if claimName == "" {
		err = fmt.Errorf("persistent volume claim name must be specified")
	} else if path == "" {
		err = fmt.Errorf("container mount path must be specified")
	} else {
		err = ValidateMountAccessFlag(access)
	}
	if err != nil {
		return nil, err
	}
	exitingPvc := openapi.NewPvcInstance()
	if access != "" {
		pvcAccessMode := pvcAccessModes(access)
		claimInfo := openapi.ClaimInfo{
			AccessModes: *openapi.NewNullablePvcAccessModes(pvcAccessMode),
		}
		exitingPvc.SetClaimInfo(claimInfo)
	}
	exitingPvc.SetExistingPvc(true)
	exitingPvc.SetClaimName(claimName)
	exitingPvc.SetPath(path)

	return exitingPvc, nil
}

func pvcAccessModes(access string) *openapi.PvcAccessModes {
	pvcAccessMode := &openapi.PvcAccessModes{}
	if access == AccessModeReadWriteParam {
		pvcAccessMode.ReadWriteMany = *openapi.NewNullableBool(ptr.To(true))
	} else {
		pvcAccessMode.ReadOnlyMany = *openapi.NewNullableBool(ptr.To(true))
	}
	return pvcAccessMode
}

// Assumption: len(pvcParams) == 3 || len(pvcParams) == 4
func handleDynamicProvisionPvc(pvcParams []string) (*openapi.PvcInstance, error) {
	var err error
	if len(pvcParams) == ThreeParams {
		// Always normalize to len() == 4 to make it easier for us in the helm chart
		pvcParams = append(pvcParams, "")
	}
	pvcClassName, storageSize, mountPath, access := pvcParams[0], pvcParams[1], pvcParams[2], pvcParams[3]
	if storageSize == "" {
		return nil, fmt.Errorf("persistent volume size must be specified")
	} else if mountPath == "" {
		return nil, fmt.Errorf("container mount path must be specified")
	} else if err = ValidateStorageClassName(pvcClassName); err != nil {
		return nil, err
	} else if err = ValidateStorageResourceRequest(storageSize); err != nil {
		return nil, err
	} else if err = ValidateMountAccessFlag(access); err != nil {
		return nil, err
	}

	pvcAccessMode := pvcAccessModes(access)
	claimInfo := openapi.ClaimInfo{
		Size:        *openapi.NewNullableString(ptr.To(storageSize)),
		AccessModes: *openapi.NewNullablePvcAccessModes(pvcAccessMode),
	}
	if pvcClassName != "" {
		claimInfo.SetStorageClass(pvcClassName)
	}
	newPvc := openapi.NewPvcInstance()
	newPvc.SetPath(mountPath)
	newPvc.SetExistingPvc(false)
	newPvc.SetClaimInfo(claimInfo)
	return newPvc, nil
}

func ValidateStorageClassName(storageClassName string) error {
	if storageClassName != "" && !dns1123LabelRegexp.MatchString(storageClassName) {
		return fmt.Errorf("A Storage Class name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character") //nolint:staticcheck // this is user facing message
	}
	return nil
}

func ValidateMountAccessFlag(accessFlag string) error {
	if accessFlag != "" && accessFlag != AccessModeReadOnlyParam && accessFlag != AccessModeReadWriteParam {
		return fmt.Errorf("invalid access permission parameter given: '%s'. this parameter may either be empty, 'ro' or 'rw'", accessFlag)
	}
	return nil
}

func ValidateStorageResourceRequest(resourceRequest string) error {
	if !resourceRequestRegex.MatchString(resourceRequest) {
		return fmt.Errorf("Badly formatted resource request for volume size requierment.\nYou can read on how to request storage resorces here: " + //nolint:staticcheck // this is user facing message
			"https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#setting-requests-and-limits-for-local-ephemeral-storage")
	}
	return nil
}
