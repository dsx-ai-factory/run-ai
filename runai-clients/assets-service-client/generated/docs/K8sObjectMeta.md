# K8sObjectMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to **map[string]string** | Free-form annotations attached to the resource. Keys must conform to Kubernetes qualified-name syntax (https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/#syntax-and-character-set): an optional DNS-subdomain prefix (≤253 chars) followed by \&quot;/\&quot; and a required name component ([A-Za-z0-9] with &#x60;.&#x60;, &#x60;_&#x60;, &#x60;-&#x60; allowed inside, ≤63 chars). Key syntax is enforced at runtime on both create and update; annotation values have no format constraint beyond the length limit below.  | [optional] 
**Labels** | Pointer to **map[string]string** | Free-form labels attached to the resource. Keys and values must conform to Kubernetes label syntax (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set): keys are qualified names (optional DNS-subdomain prefix ≤253 chars + \&quot;/\&quot; + name component ≤63 chars); values are ≤63 chars matching the same [A-Za-z0-9]/&#x60;.&#x60;/&#x60;_&#x60;/&#x60;-&#x60; character set (empty values are allowed). Key syntax is validated at runtime; the value pattern below is enforced by the schema as well as at runtime.  | [optional] 

## Methods

### NewK8sObjectMeta

`func NewK8sObjectMeta() *K8sObjectMeta`

NewK8sObjectMeta instantiates a new K8sObjectMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewK8sObjectMetaWithDefaults

`func NewK8sObjectMetaWithDefaults() *K8sObjectMeta`

NewK8sObjectMetaWithDefaults instantiates a new K8sObjectMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *K8sObjectMeta) GetAnnotations() map[string]string`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *K8sObjectMeta) GetAnnotationsOk() (*map[string]string, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *K8sObjectMeta) SetAnnotations(v map[string]string)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *K8sObjectMeta) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *K8sObjectMeta) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *K8sObjectMeta) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *K8sObjectMeta) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *K8sObjectMeta) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *K8sObjectMeta) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *K8sObjectMeta) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *K8sObjectMeta) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *K8sObjectMeta) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


