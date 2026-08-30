// Package v1alpha1 contains local NIMService types with MultiNode support
// This extends the upstream nim-operator types to include fields not present in v1.0.0
package v1alpha1

import (
	"fmt"
	"strings"

	nim "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Re-export status constants from upstream
const (
	NIMServiceConditionReady  = nim.NIMServiceConditionReady
	NIMServiceConditionFailed = nim.NIMServiceConditionFailed
	NIMServiceStatusPending   = nim.NIMServiceStatusPending
	NIMServiceStatusNotReady  = nim.NIMServiceStatusNotReady
	NIMServiceStatusReady     = nim.NIMServiceStatusReady
	NIMServiceStatusFailed    = nim.NIMServiceStatusFailed
)

// --- CEL utility types (local, avoids dependency on k8s-nim-operator/internal/k8sutil/cel) ---

// ComparisonOperator defines available operators for CEL expression.
type ComparisonOperator string

const (
	OpEqual          ComparisonOperator = "=="
	OpNotEqual       ComparisonOperator = "!="
	OpGreater        ComparisonOperator = ">"
	OpGreaterOrEqual ComparisonOperator = ">="
	OpLess           ComparisonOperator = "<"
	OpLessOrEqual    ComparisonOperator = "<="
)

// ValueType defines available value types for CEL.
type ValueType string

const (
	TypeBool     ValueType = "bool"
	TypeString   ValueType = "string"
	TypeInt      ValueType = "int"
	TypeSemver   ValueType = "semver"
	TypeQuantity ValueType = "quantity"
	TypeUnknown  ValueType = "unknown"
)

// splitQualifiedName splits a qualified name into domain and name.
// If the name is not qualified, the defaultDomain is used.
func splitQualifiedName(name string, defaultDomain string) (string, string) {
	sep := strings.Index(name, "/")
	if sep == -1 {
		return defaultDomain, name
	}
	return name[0:sep], name[sep+1:]
}

var boolOperators = map[ComparisonOperator]string{
	OpEqual:    "==",
	OpNotEqual: "!=",
}

var intOperators = map[ComparisonOperator]string{
	OpEqual:          "==",
	OpNotEqual:       "!=",
	OpGreater:        ">",
	OpGreaterOrEqual: ">=",
	OpLess:           "<",
	OpLessOrEqual:    "<=",
}

var stringOperators = map[ComparisonOperator]string{
	OpEqual:    "==",
	OpNotEqual: "!=",
}

var semverOperators = map[ComparisonOperator]string{
	OpEqual:          "== 0",
	OpNotEqual:       "!= 0",
	OpGreater:        "> 0",
	OpGreaterOrEqual: ">= 0",
	OpLess:           "< 0",
	OpLessOrEqual:    "<= 0",
}

var quantityOperators = map[ComparisonOperator]string{
	OpEqual:          "== 0",
	OpNotEqual:       "!= 0",
	OpGreater:        "> 0",
	OpGreaterOrEqual: ">= 0",
	OpLess:           "< 0",
	OpLessOrEqual:    "<= 0",
}

// buildCELExpr returns a CEL expression given key, operator, value, and type.
func buildCELExpr(key string, op ComparisonOperator, value interface{}, vt ValueType) (string, error) {
	if value == nil {
		return "", fmt.Errorf("value is nil")
	}

	var expr string
	switch vt {
	case TypeBool:
		if b, ok := value.(bool); ok {
			comp, ok := boolOperators[op]
			if !ok {
				return "", fmt.Errorf("invalid operator %q for bool type", op)
			}
			expr = fmt.Sprintf("%s %s %t", key, comp, b)
		}
	case TypeInt:
		if num, ok := value.(int); ok {
			comp, ok := intOperators[op]
			if !ok {
				return "", fmt.Errorf("invalid operator %q for int type", op)
			}
			expr = fmt.Sprintf("%s %s %d", key, comp, num)
		}
	case TypeString:
		if str, ok := value.(string); ok {
			comp, ok := stringOperators[op]
			if !ok {
				return "", fmt.Errorf("invalid operator %q for string type", op)
			}
			expr = fmt.Sprintf("%s %s %q", key, comp, str)
		}
	case TypeSemver:
		if str, ok := value.(string); ok {
			comp, ok := semverOperators[op]
			if !ok {
				return "", fmt.Errorf("invalid operator %q for semver type", op)
			}
			expr = fmt.Sprintf("(%s).compareTo(semver(%q)) %s", key, str, comp)
		}
	case TypeQuantity:
		if quantity, ok := value.(*apiresource.Quantity); ok {
			comp, ok := quantityOperators[op]
			if !ok {
				return "", fmt.Errorf("invalid operator %q for quantity type", op)
			}
			expr = fmt.Sprintf("(%s).compareTo(quantity(%q)) %s", key, quantity.String(), comp)
		}
	}
	if expr == "" {
		return "", fmt.Errorf("invalid value type %q", vt)
	}
	return expr, nil
}

// --- NIMServiceSpec ---

// NIMServiceSpec defines the desired state of NIMService
// This embeds the upstream spec and adds MultiNode support
type NIMServiceSpec struct {
	// Embed the upstream NIMServiceSpec
	nim.NIMServiceSpec `json:",inline"`

	// MultiNode configuration for distributed NIM deployments
	// This field is not present in nim-operator v1.0.0 but is needed for multi-node support
	MultiNode *MultiNodeConfig `json:"multiNode,omitempty"`

	DRAResources []DRAResource `json:"draResources,omitempty"`
}

// --- DRA types ---

// DRAResource references exactly one ResourceClaim, either directly
// or by naming a ResourceClaimTemplate which is then turned into a ResourceClaim.
//
// When creating the NIMService pods, it adds a name (`DNS_LABEL` format) to it
// that uniquely identifies the DRA resource.
// +kubebuilder:validation:XValidation:rule="(has(self.resourceClaimName) ? 1 : 0) + (has(self.resourceClaimTemplateName) ? 1 : 0) + (has(self.claimCreationSpec) ? 1 : 0) == 1",message="exactly one of spec.resourceClaimName, spec.resourceClaimTemplateName, or spec.claimCreationSpec must be set."
type DRAResource struct {
	// ResourceClaimName is the name of a DRA resource claim object in the same
	// namespace as the NIMService.
	//
	// Exactly one of ResourceClaimName and ResourceClaimTemplateName must
	// be set.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	ResourceClaimName *string `json:"resourceClaimName,omitempty"`

	// ResourceClaimTemplateName is the name of a DRA resource claim template
	// object in the same namespace as the pods for this NIMService.
	//
	// The template will be used to create a new DRA resource claim, which will
	// be bound to the pods created for this NIMService.
	//
	// Exactly one of ResourceClaimName and ResourceClaimTemplateName must
	// be set.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	ResourceClaimTemplateName *string `json:"resourceClaimTemplateName,omitempty"`

	// ClaimCreationSpec is the spec to auto-generate a DRA resource claim template.
	// Only one of ClaimCreationSpec, ResourceClaimName or ResourceClaimTemplateName must be specified.
	ClaimCreationSpec *DRAClaimCreationSpec `json:"claimCreationSpec,omitempty"`

	// Requests is the list of requests in the referenced DRA resource claim.
	// to be made available to the model container of the NIMService pods.
	//
	// If empty, everything from the claim is made available, otherwise
	// only the result of this subset of requests.
	//
	// +kubebuilder:validation:items:MinLength=1
	Requests []string `json:"requests,omitempty"`
}

// DRADeviceAttributeSelectorValue defines the value of a device attribute to compare.
// Exactly one of the fields must be set.
type DRADeviceAttributeSelectorValue struct {
	// BoolValue is a true/false value.
	BoolValue *bool `json:"boolValue,omitempty"`
	// IntValue is a number.
	IntValue *int32 `json:"intValue,omitempty"`
	// StringValue is a string value.
	// +kubebuilder:validation:MaxLength=64
	StringValue *string `json:"stringValue,omitempty"`
	// VersionValue is a semantic version according to semver.org spec 2.0.0.
	// +kubebuilder:validation:MaxLength=64
	VersionValue *string `json:"versionValue,omitempty"`
}

func (d *DRADeviceAttributeSelectorValue) GetValue() any {
	switch {
	case d.BoolValue != nil:
		return *d.BoolValue
	case d.IntValue != nil:
		return int(*d.IntValue)
	case d.StringValue != nil:
		return *d.StringValue
	case d.VersionValue != nil:
		return *d.VersionValue
	}
	return nil
}

func (d *DRADeviceAttributeSelectorValue) GetValueType() ValueType {
	switch {
	case d.BoolValue != nil:
		return TypeBool
	case d.IntValue != nil:
		return TypeInt
	case d.StringValue != nil:
		return TypeString
	case d.VersionValue != nil:
		return TypeSemver
	default:
		return TypeUnknown
	}
}

// DRADeviceAttributeSelectorOp defines the operator to use for comparing against a device attribute value.
type DRADeviceAttributeSelectorOp string

const (
	DRADeviceAttributeSelectorOpEqual              DRADeviceAttributeSelectorOp = "Equal"
	DRADeviceAttributeSelectorOpNotEqual           DRADeviceAttributeSelectorOp = "NotEqual"
	DRADeviceAttributeSelectorOpGreaterThan        DRADeviceAttributeSelectorOp = "GreaterThan"
	DRADeviceAttributeSelectorOpGreaterThanOrEqual DRADeviceAttributeSelectorOp = "GreaterThanOrEqual"
	DRADeviceAttributeSelectorOpLessThan           DRADeviceAttributeSelectorOp = "LessThan"
	DRADeviceAttributeSelectorOpLessThanOrEqual    DRADeviceAttributeSelectorOp = "LessThanOrEqual"
)

func (d DRADeviceAttributeSelectorOp) GetCELOperator() ComparisonOperator {
	switch d {
	case DRADeviceAttributeSelectorOpEqual:
		return OpEqual
	case DRADeviceAttributeSelectorOpNotEqual:
		return OpNotEqual
	case DRADeviceAttributeSelectorOpGreaterThan:
		return OpGreater
	case DRADeviceAttributeSelectorOpGreaterThanOrEqual:
		return OpGreaterOrEqual
	case DRADeviceAttributeSelectorOpLessThan:
		return OpLess
	case DRADeviceAttributeSelectorOpLessThanOrEqual:
		return OpLessOrEqual
	default:
		return OpEqual
	}
}

// DRADeviceAttributeSelector defines the selector expression for a DRA device attribute.
type DRADeviceAttributeSelector struct {
	// Key is the name of the device attribute.
	// This is either a qualified name or a simple name.
	// If it is a simple name, then it is assumed to be prefixed with the DRA driver name.
	// Eg: "gpu.nvidia.com/productName" is equivalent to "productName" if the driver name is "gpu.nvidia.com". Otherwise they're treated as 2 different attributes.
	// +kubebuilder:validation:MaxLength=64
	Key string `json:"key"`
	// Op is the operator to use for comparing the device attribute value. Supported operators are:
	// * Equal: The device attribute value must be equal to the value specified in the selector.
	// * NotEqual: The device attribute value must not be equal to the value specified in the selector.
	// * GreaterThan: The device attribute value must be greater than the value specified in the selector.
	// * GreaterThanOrEqual: The device attribute value must be greater than or equal to the value specified in the selector.
	// * LessThan: The device attribute value must be less than the value specified in the selector.
	// * LessThanOrEqual: The device attribute value must be less than or equal to the value specified in the selector.
	//
	// +kubebuilder:validation:Enum=Equal;NotEqual;GreaterThan;GreaterThanOrEqual;LessThan;LessThanOrEqual
	// +kubebuilder:default=Equal
	Op DRADeviceAttributeSelectorOp `json:"op"`
	// Value is the value to compare against the device attribute.
	Value *DRADeviceAttributeSelectorValue `json:"value,omitempty"`
}

func (d *DRADeviceAttributeSelector) GetCELExpression(driverName string) (string, error) {
	domain, name := splitQualifiedName(d.Key, driverName)
	attrKey := fmt.Sprintf("device.attributes[%q].%s", domain, name)
	return buildCELExpr(attrKey, d.Op.GetCELOperator(), d.Value.GetValue(), d.Value.GetValueType())
}

// DRAResourceQuantitySelectorOp defines the operator to use for comparing against a resource quantity.
type DRAResourceQuantitySelectorOp string

const (
	DRAResourceQuantitySelectorOpEqual              DRAResourceQuantitySelectorOp = "Equal"
	DRAResourceQuantitySelectorOpNotEqual           DRAResourceQuantitySelectorOp = "NotEqual"
	DRAResourceQuantitySelectorOpGreaterThan        DRAResourceQuantitySelectorOp = "GreaterThan"
	DRAResourceQuantitySelectorOpGreaterThanOrEqual DRAResourceQuantitySelectorOp = "GreaterThanOrEqual"
	DRAResourceQuantitySelectorOpLessThan           DRAResourceQuantitySelectorOp = "LessThan"
	DRAResourceQuantitySelectorOpLessThanOrEqual    DRAResourceQuantitySelectorOp = "LessThanOrEqual"
)

func (d DRAResourceQuantitySelectorOp) GetCELOperator() ComparisonOperator {
	switch d {
	case DRAResourceQuantitySelectorOpEqual:
		return OpEqual
	case DRAResourceQuantitySelectorOpNotEqual:
		return OpNotEqual
	case DRAResourceQuantitySelectorOpGreaterThan:
		return OpGreater
	case DRAResourceQuantitySelectorOpGreaterThanOrEqual:
		return OpGreaterOrEqual
	case DRAResourceQuantitySelectorOpLessThan:
		return OpLess
	case DRAResourceQuantitySelectorOpLessThanOrEqual:
		return OpLessOrEqual
	default:
		return OpEqual
	}
}

// DRAResourceQuantitySelector defines the selector expression for a DRA device capacity.
type DRAResourceQuantitySelector struct {
	// Key is the name of the resource.
	// This is either a qualified name or a simple name.
	// If it is a simple name, then it is assumed to be prefixed with the DRA driver name.
	// Eg: "gpu.nvidia.com/memory" is equivalent to "memory" if the driver name is "gpu.nvidia.com". Otherwise they're treated as 2 different attributes.
	// +kubebuilder:validation:MaxLength=64
	Key string `json:"key"`
	// Op is the operator to use for comparing against the device capacity. Supported operators are:
	// * Equal: The resource quantity value must be equal to the value specified in the selector.
	// * NotEqual: The resource quantity value must not be equal to the value specified in the selector.
	// * GreaterThan: The resource quantity value must be greater than the value specified in the selector.
	// * GreaterThanOrEqual: The resource quantity value must be greater than or equal to the value specified in the selector.
	// * LessThan: The resource quantity value must be less than the value specified in the selector.
	// * LessThanOrEqual: The resource quantity value must be less than or equal to the value specified in the selector.
	//
	// +kubebuilder:validation:Enum=Equal;NotEqual;GreaterThan;GreaterThanOrEqual;LessThan;LessThanOrEqual
	// +kubebuilder:default=Equal
	Op DRAResourceQuantitySelectorOp `json:"op"`
	// Value is the resource quantity to compare against.
	//
	// +kubebuilder:validation:Required
	Value *apiresource.Quantity `json:"value,omitempty"`
}

func (d *DRAResourceQuantitySelector) GetCELExpression(driverName string) (string, error) {
	domain, name := splitQualifiedName(d.Key, driverName)
	attrKey := fmt.Sprintf("device.capacity[%q].%s", domain, name)
	return buildCELExpr(attrKey, d.Op.GetCELOperator(), d.Value, TypeQuantity)
}

// +kubebuilder:validation:XValidation:rule="(has(self.celExpressions) && !(has(self.attributeSelectors) || has(self.capacitySelectors))) || !has(self.celExpressions)",message="celExpressions must not be set if attributeSelectors or capacitySelectors are set"
type DRADeviceSpec struct {
	// Name is the name of the device request to use in the generated claim spec.
	// Must be a valid DNS_LABEL.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	Name string `json:"name"`
	// Count is the number of devices to request.
	// +kubebuilder:default=1
	Count uint32 `json:"count"`
	// DeviceClassName references a specific DeviceClass to inherit configuration and selectors from.
	// +kubebuilder:default=gpu.nvidia.com
	DeviceClassName string `json:"deviceClassName"`
	// DriverName is the name of the DRA driver providing the capacity information.
	// Must be a DNS subdomain.
	//
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	// +kubebuilder:default=gpu.nvidia.com
	DriverName string `json:"driverName,omitempty"`
	// AttributeSelectors defines the criteria which must be satisfied by the device attributes of a device.
	// +kubebuilder:validation:MaxSize=20
	AttributeSelectors []DRADeviceAttributeSelector `json:"attributeSelectors,omitempty"`
	// CapacitySelectors defines the criteria which must be satisfied by the device capacity of a device.
	// +kubebuilder:validation:MaxSize=12
	CapacitySelectors []DRAResourceQuantitySelector `json:"capacitySelectors,omitempty"`
	// CELExpressions is a list of CEL expressions that must be satisfied by the DRA device.
	CELExpressions []string `json:"celExpressions,omitempty"`
}

// DRAClaimCreationSpec defines the spec for generating a DRA resource claim template.
type DRAClaimCreationSpec struct {
	// GenerateName is an optional name prefix to use for generating the resource claim template.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=16
	GenerateName string `json:"generateName,omitempty"`
	// +kubebuilder:validation:MinSize=1
	Devices []DRADeviceSpec `json:"devices"`
}

func (d *DRAClaimCreationSpec) GetNamePrefix() string {
	namePrefix := d.GenerateName
	if namePrefix != "" {
		return namePrefix
	}
	return "claimtemplate"
}

// DRAResourceStatus defines the status of the DRAResource.
// +kubebuilder:validation:XValidation:rule="has(self.resourceClaimStatus) != has(self.resourceClaimTemplateStatus)",message="exactly one of resourceClaimStatus and resourceClaimTemplateStatus must be set."
type DRAResourceStatus struct {
	// Name is the pod claim name referenced in the pod spec as `spec.resourceClaims[].name` for this DRA resource.
	Name string `json:"name"`
	// ResourceClaimStatus is the status of the resource claim in this DRA resource.
	//
	// Exactly one of resourceClaimStatus and resourceClaimTemplateStatus will be set.
	ResourceClaimStatus *DRAResourceClaimStatusInfo `json:"resourceClaimStatus,omitempty"`
	// ResourceClaimTemplateStatus is the status of the resource claim template in this DRA resource.
	//
	// Exactly one of resourceClaimStatus and resourceClaimTemplateStatus will be set.
	ResourceClaimTemplateStatus *DRAResourceClaimTemplateStatusInfo `json:"resourceClaimTemplateStatus,omitempty"`
}

// DRAResourceClaimStatusInfo defines the status of a ResourceClaim referenced in the DRAResource.
type DRAResourceClaimStatusInfo struct {
	// Name is the name of the ResourceClaim.
	Name string `json:"name"`
	// State is the state of the ResourceClaim.
	// * pending: the resource claim is pending allocation.
	// * deleted: the resource claim has a deletion timestamp set but is not yet finalized.
	// * allocated: the resource claim is allocated to a pod.
	// * reserved: the resource claim is consumed by a pod.
	// This field will have one or more of the above values depending on the status of the resource claim.
	//
	// +kubebuilder:validation:default=pending
	State string `json:"state"`
}

// DRAResourceClaimTemplateStatusInfo defines the status of a ResourceClaimTemplate referenced in the DRAResource.
type DRAResourceClaimTemplateStatusInfo struct {
	// Name is the name of the resource claim template.
	Name string `json:"name"`
	// ResourceClaimStatuses is the statuses of the generated resource claims from this resource claim template.
	ResourceClaimStatuses []DRAResourceClaimStatusInfo `json:"resourceClaimStatuses,omitempty"`
}

// --- MultiNode types ---

// MultiNodeConfig defines the multi-node configuration for distributed NIM
type MultiNodeConfig struct {
	// Parallelism defines the parallelism configuration
	Parallelism *ParallelismSpec `json:"parallelism,omitempty"`
}

// ParallelismSpec defines parallelism settings for multi-node deployment
type ParallelismSpec struct {
	// Pipeline is the number of nodes per replica (1 leader + N workers)
	// For example, pipeline=4 means 1 leader + 3 workers = 4 nodes per replica
	Pipeline *int `json:"pipeline,omitempty"`
	// Tensor is the tensor parallelism setting
	Tensor *int `json:"tensor,omitempty"`
}

// --- NIMService CRD ---

// NIMService is the Schema for the nimservices API
// This version includes MultiNode support
type NIMService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NIMServiceSpec       `json:"spec,omitempty"`
	Status nim.NIMServiceStatus `json:"status,omitempty"`
}

// GetPipelineParallelism returns the pipeline parallelism value, or 1 if not set
func (n *NIMService) GetPipelineParallelism() int {
	if n.Spec.MultiNode == nil || n.Spec.MultiNode.Parallelism == nil || n.Spec.MultiNode.Parallelism.Pipeline == nil {
		return 1
	}
	if *n.Spec.MultiNode.Parallelism.Pipeline < 1 {
		return 1
	}
	return *n.Spec.MultiNode.Parallelism.Pipeline
}

// NIMServiceList contains a list of NIMService
type NIMServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NIMService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NIMService{}, &NIMServiceList{})
}
