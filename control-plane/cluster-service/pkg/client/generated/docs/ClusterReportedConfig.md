# ClusterReportedConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InferenceEndpointSchemaVersion** | Pointer to **NullableString** | InferenceEndpoint integration contract declared by the operator when Agent, Anyworkload and inference-endpoint-controller are enabled. Omitted by older clusters. Operand readiness and administrator Gateway settings must also be checked before enabling new workloads. | [optional] 
**WorkloadOwnershipProtection** | Pointer to **NullableBool** | indicates if the cluster workload ownership protection is enabled | [optional] 
**SubdomainEnabled** | Pointer to **NullableBool** | indicates if the cluster subdomain is enabled | [optional] 
**IngressClass** | Pointer to **NullableString** | the ingress class that is configured for the cluster | [optional] 
**IngressEnabled** | Pointer to **NullableBool** | indicates if Ingress or Route resources should be created for workload external URLs | [optional] 
**GatewayName** | Pointer to **NullableString** | the name of the Gateway API gateway resource configured for the cluster | [optional] 
**GatewayNamespace** | Pointer to **NullableString** | the namespace of the Gateway API gateway resource configured for the cluster | [optional] 
**GatewayAPIEnabled** | Pointer to **NullableBool** | indicates if Gateway API HTTPRoute resources should be created for workload external URLs | [optional] 
**GitSyncImage** | Pointer to **NullableString** | the fully qualified git-sync image reference used by init containers (format \&quot;registry/name:tag\&quot;) | [optional] 
**GitSyncResources** | Pointer to [**NullableClusterReportedResourceRequirements**](ClusterReportedResourceRequirements.md) |  | [optional] 
**S3FileSystemImage** | Pointer to **NullableString** | the fully qualified s3 file-system image reference used by init containers (format \&quot;registry/name:tag\&quot;) | [optional] 
**S3FileSystemResources** | Pointer to [**NullableClusterReportedResourceRequirements**](ClusterReportedResourceRequirements.md) |  | [optional] 
**FipsMode** | Pointer to [**NullableFipsModeEnum**](FipsModeEnum.md) |  | [optional] 
**RunaiOauth2ProxyImage** | Pointer to **NullableString** | the fully qualified runai oauth2-proxy image reference used by init containers (format \&quot;registry/name:tag\&quot;) | [optional] 
**RunaiOauth2ProxyResources** | Pointer to [**NullableClusterReportedResourceRequirements**](ClusterReportedResourceRequirements.md) |  | [optional] 
**RunaiOauth2ProxyUsernameLoaderImage** | Pointer to **NullableString** | the fully qualified runai oauth2-proxy username loader image reference used by init containers (format \&quot;registry/name:tag\&quot;) | [optional] 
**UsernameLoaderResources** | Pointer to [**NullableClusterReportedResourceRequirements**](ClusterReportedResourceRequirements.md) |  | [optional] 
**RunaiOauth2ProxyRedisFQDN** | Pointer to **NullableString** | the FQDN of the Redis backing oauth2-proxy | [optional] 
**RunaiOauth2ProxyWorkloadAuthSecretName** | Pointer to **NullableString** | the name of the Secret holding the oauth2-proxy workload auth credentials | [optional] 
**RunaiOauth2ProxySessionStoreType** | Pointer to **NullableString** | the session store type used by oauth2-proxy | [optional] 
**RunaiOauth2ProxyOidcIssuerUrl** | Pointer to **NullableString** | the absolute OIDC issuer URL oauth2-proxy uses to discover the OpenID provider | [optional] 
**CustomCAGit** | Pointer to [**NullableClusterReportedCustomCA**](ClusterReportedCustomCA.md) |  | [optional] 
**CustomCAS3** | Pointer to [**NullableClusterReportedCustomCA**](ClusterReportedCustomCA.md) |  | [optional] 
**CustomCAOauth2Proxy** | Pointer to [**NullableClusterReportedCustomCA**](ClusterReportedCustomCA.md) |  | [optional] 
**ExternalAuthUrlEnabled** | Pointer to **NullableBool** | indicates if OAuth2 proxy is used when submitting workloads with an external URL | [optional] 
**OpenshiftRouteTerminationPolicy** | Pointer to [**NullableOpenshiftRouteTerminationPolicy**](OpenshiftRouteTerminationPolicy.md) |  | [optional] 
**DistGPUNetworkAcceleration** | Pointer to **NullableBool** | indicates if GPU network acceleration (MNNVL) is enabled for the workload-controller (distributed-training) | [optional] 
**DinfGPUNetworkAcceleration** | Pointer to **NullableBool** | indicates if GPU network acceleration (MNNVL) is enabled for the inference-workload-controller (distributed-inference) | [optional] 

## Methods

### NewClusterReportedConfig

`func NewClusterReportedConfig() *ClusterReportedConfig`

NewClusterReportedConfig instantiates a new ClusterReportedConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedConfigWithDefaults

`func NewClusterReportedConfigWithDefaults() *ClusterReportedConfig`

NewClusterReportedConfigWithDefaults instantiates a new ClusterReportedConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInferenceEndpointSchemaVersion

`func (o *ClusterReportedConfig) GetInferenceEndpointSchemaVersion() string`

GetInferenceEndpointSchemaVersion returns the InferenceEndpointSchemaVersion field if non-nil, zero value otherwise.

### GetInferenceEndpointSchemaVersionOk

`func (o *ClusterReportedConfig) GetInferenceEndpointSchemaVersionOk() (*string, bool)`

GetInferenceEndpointSchemaVersionOk returns a tuple with the InferenceEndpointSchemaVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferenceEndpointSchemaVersion

`func (o *ClusterReportedConfig) SetInferenceEndpointSchemaVersion(v string)`

SetInferenceEndpointSchemaVersion sets InferenceEndpointSchemaVersion field to given value.

### HasInferenceEndpointSchemaVersion

`func (o *ClusterReportedConfig) HasInferenceEndpointSchemaVersion() bool`

HasInferenceEndpointSchemaVersion returns a boolean if a field has been set.

### SetInferenceEndpointSchemaVersionNil

`func (o *ClusterReportedConfig) SetInferenceEndpointSchemaVersionNil(b bool)`

 SetInferenceEndpointSchemaVersionNil sets the value for InferenceEndpointSchemaVersion to be an explicit nil

### UnsetInferenceEndpointSchemaVersion
`func (o *ClusterReportedConfig) UnsetInferenceEndpointSchemaVersion()`

UnsetInferenceEndpointSchemaVersion ensures that no value is present for InferenceEndpointSchemaVersion, not even an explicit nil
### GetWorkloadOwnershipProtection

`func (o *ClusterReportedConfig) GetWorkloadOwnershipProtection() bool`

GetWorkloadOwnershipProtection returns the WorkloadOwnershipProtection field if non-nil, zero value otherwise.

### GetWorkloadOwnershipProtectionOk

`func (o *ClusterReportedConfig) GetWorkloadOwnershipProtectionOk() (*bool, bool)`

GetWorkloadOwnershipProtectionOk returns a tuple with the WorkloadOwnershipProtection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadOwnershipProtection

`func (o *ClusterReportedConfig) SetWorkloadOwnershipProtection(v bool)`

SetWorkloadOwnershipProtection sets WorkloadOwnershipProtection field to given value.

### HasWorkloadOwnershipProtection

`func (o *ClusterReportedConfig) HasWorkloadOwnershipProtection() bool`

HasWorkloadOwnershipProtection returns a boolean if a field has been set.

### SetWorkloadOwnershipProtectionNil

`func (o *ClusterReportedConfig) SetWorkloadOwnershipProtectionNil(b bool)`

 SetWorkloadOwnershipProtectionNil sets the value for WorkloadOwnershipProtection to be an explicit nil

### UnsetWorkloadOwnershipProtection
`func (o *ClusterReportedConfig) UnsetWorkloadOwnershipProtection()`

UnsetWorkloadOwnershipProtection ensures that no value is present for WorkloadOwnershipProtection, not even an explicit nil
### GetSubdomainEnabled

`func (o *ClusterReportedConfig) GetSubdomainEnabled() bool`

GetSubdomainEnabled returns the SubdomainEnabled field if non-nil, zero value otherwise.

### GetSubdomainEnabledOk

`func (o *ClusterReportedConfig) GetSubdomainEnabledOk() (*bool, bool)`

GetSubdomainEnabledOk returns a tuple with the SubdomainEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubdomainEnabled

`func (o *ClusterReportedConfig) SetSubdomainEnabled(v bool)`

SetSubdomainEnabled sets SubdomainEnabled field to given value.

### HasSubdomainEnabled

`func (o *ClusterReportedConfig) HasSubdomainEnabled() bool`

HasSubdomainEnabled returns a boolean if a field has been set.

### SetSubdomainEnabledNil

`func (o *ClusterReportedConfig) SetSubdomainEnabledNil(b bool)`

 SetSubdomainEnabledNil sets the value for SubdomainEnabled to be an explicit nil

### UnsetSubdomainEnabled
`func (o *ClusterReportedConfig) UnsetSubdomainEnabled()`

UnsetSubdomainEnabled ensures that no value is present for SubdomainEnabled, not even an explicit nil
### GetIngressClass

`func (o *ClusterReportedConfig) GetIngressClass() string`

GetIngressClass returns the IngressClass field if non-nil, zero value otherwise.

### GetIngressClassOk

`func (o *ClusterReportedConfig) GetIngressClassOk() (*string, bool)`

GetIngressClassOk returns a tuple with the IngressClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIngressClass

`func (o *ClusterReportedConfig) SetIngressClass(v string)`

SetIngressClass sets IngressClass field to given value.

### HasIngressClass

`func (o *ClusterReportedConfig) HasIngressClass() bool`

HasIngressClass returns a boolean if a field has been set.

### SetIngressClassNil

`func (o *ClusterReportedConfig) SetIngressClassNil(b bool)`

 SetIngressClassNil sets the value for IngressClass to be an explicit nil

### UnsetIngressClass
`func (o *ClusterReportedConfig) UnsetIngressClass()`

UnsetIngressClass ensures that no value is present for IngressClass, not even an explicit nil
### GetIngressEnabled

`func (o *ClusterReportedConfig) GetIngressEnabled() bool`

GetIngressEnabled returns the IngressEnabled field if non-nil, zero value otherwise.

### GetIngressEnabledOk

`func (o *ClusterReportedConfig) GetIngressEnabledOk() (*bool, bool)`

GetIngressEnabledOk returns a tuple with the IngressEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIngressEnabled

`func (o *ClusterReportedConfig) SetIngressEnabled(v bool)`

SetIngressEnabled sets IngressEnabled field to given value.

### HasIngressEnabled

`func (o *ClusterReportedConfig) HasIngressEnabled() bool`

HasIngressEnabled returns a boolean if a field has been set.

### SetIngressEnabledNil

`func (o *ClusterReportedConfig) SetIngressEnabledNil(b bool)`

 SetIngressEnabledNil sets the value for IngressEnabled to be an explicit nil

### UnsetIngressEnabled
`func (o *ClusterReportedConfig) UnsetIngressEnabled()`

UnsetIngressEnabled ensures that no value is present for IngressEnabled, not even an explicit nil
### GetGatewayName

`func (o *ClusterReportedConfig) GetGatewayName() string`

GetGatewayName returns the GatewayName field if non-nil, zero value otherwise.

### GetGatewayNameOk

`func (o *ClusterReportedConfig) GetGatewayNameOk() (*string, bool)`

GetGatewayNameOk returns a tuple with the GatewayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGatewayName

`func (o *ClusterReportedConfig) SetGatewayName(v string)`

SetGatewayName sets GatewayName field to given value.

### HasGatewayName

`func (o *ClusterReportedConfig) HasGatewayName() bool`

HasGatewayName returns a boolean if a field has been set.

### SetGatewayNameNil

`func (o *ClusterReportedConfig) SetGatewayNameNil(b bool)`

 SetGatewayNameNil sets the value for GatewayName to be an explicit nil

### UnsetGatewayName
`func (o *ClusterReportedConfig) UnsetGatewayName()`

UnsetGatewayName ensures that no value is present for GatewayName, not even an explicit nil
### GetGatewayNamespace

`func (o *ClusterReportedConfig) GetGatewayNamespace() string`

GetGatewayNamespace returns the GatewayNamespace field if non-nil, zero value otherwise.

### GetGatewayNamespaceOk

`func (o *ClusterReportedConfig) GetGatewayNamespaceOk() (*string, bool)`

GetGatewayNamespaceOk returns a tuple with the GatewayNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGatewayNamespace

`func (o *ClusterReportedConfig) SetGatewayNamespace(v string)`

SetGatewayNamespace sets GatewayNamespace field to given value.

### HasGatewayNamespace

`func (o *ClusterReportedConfig) HasGatewayNamespace() bool`

HasGatewayNamespace returns a boolean if a field has been set.

### SetGatewayNamespaceNil

`func (o *ClusterReportedConfig) SetGatewayNamespaceNil(b bool)`

 SetGatewayNamespaceNil sets the value for GatewayNamespace to be an explicit nil

### UnsetGatewayNamespace
`func (o *ClusterReportedConfig) UnsetGatewayNamespace()`

UnsetGatewayNamespace ensures that no value is present for GatewayNamespace, not even an explicit nil
### GetGatewayAPIEnabled

`func (o *ClusterReportedConfig) GetGatewayAPIEnabled() bool`

GetGatewayAPIEnabled returns the GatewayAPIEnabled field if non-nil, zero value otherwise.

### GetGatewayAPIEnabledOk

`func (o *ClusterReportedConfig) GetGatewayAPIEnabledOk() (*bool, bool)`

GetGatewayAPIEnabledOk returns a tuple with the GatewayAPIEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGatewayAPIEnabled

`func (o *ClusterReportedConfig) SetGatewayAPIEnabled(v bool)`

SetGatewayAPIEnabled sets GatewayAPIEnabled field to given value.

### HasGatewayAPIEnabled

`func (o *ClusterReportedConfig) HasGatewayAPIEnabled() bool`

HasGatewayAPIEnabled returns a boolean if a field has been set.

### SetGatewayAPIEnabledNil

`func (o *ClusterReportedConfig) SetGatewayAPIEnabledNil(b bool)`

 SetGatewayAPIEnabledNil sets the value for GatewayAPIEnabled to be an explicit nil

### UnsetGatewayAPIEnabled
`func (o *ClusterReportedConfig) UnsetGatewayAPIEnabled()`

UnsetGatewayAPIEnabled ensures that no value is present for GatewayAPIEnabled, not even an explicit nil
### GetGitSyncImage

`func (o *ClusterReportedConfig) GetGitSyncImage() string`

GetGitSyncImage returns the GitSyncImage field if non-nil, zero value otherwise.

### GetGitSyncImageOk

`func (o *ClusterReportedConfig) GetGitSyncImageOk() (*string, bool)`

GetGitSyncImageOk returns a tuple with the GitSyncImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitSyncImage

`func (o *ClusterReportedConfig) SetGitSyncImage(v string)`

SetGitSyncImage sets GitSyncImage field to given value.

### HasGitSyncImage

`func (o *ClusterReportedConfig) HasGitSyncImage() bool`

HasGitSyncImage returns a boolean if a field has been set.

### SetGitSyncImageNil

`func (o *ClusterReportedConfig) SetGitSyncImageNil(b bool)`

 SetGitSyncImageNil sets the value for GitSyncImage to be an explicit nil

### UnsetGitSyncImage
`func (o *ClusterReportedConfig) UnsetGitSyncImage()`

UnsetGitSyncImage ensures that no value is present for GitSyncImage, not even an explicit nil
### GetGitSyncResources

`func (o *ClusterReportedConfig) GetGitSyncResources() ClusterReportedResourceRequirements`

GetGitSyncResources returns the GitSyncResources field if non-nil, zero value otherwise.

### GetGitSyncResourcesOk

`func (o *ClusterReportedConfig) GetGitSyncResourcesOk() (*ClusterReportedResourceRequirements, bool)`

GetGitSyncResourcesOk returns a tuple with the GitSyncResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitSyncResources

`func (o *ClusterReportedConfig) SetGitSyncResources(v ClusterReportedResourceRequirements)`

SetGitSyncResources sets GitSyncResources field to given value.

### HasGitSyncResources

`func (o *ClusterReportedConfig) HasGitSyncResources() bool`

HasGitSyncResources returns a boolean if a field has been set.

### SetGitSyncResourcesNil

`func (o *ClusterReportedConfig) SetGitSyncResourcesNil(b bool)`

 SetGitSyncResourcesNil sets the value for GitSyncResources to be an explicit nil

### UnsetGitSyncResources
`func (o *ClusterReportedConfig) UnsetGitSyncResources()`

UnsetGitSyncResources ensures that no value is present for GitSyncResources, not even an explicit nil
### GetS3FileSystemImage

`func (o *ClusterReportedConfig) GetS3FileSystemImage() string`

GetS3FileSystemImage returns the S3FileSystemImage field if non-nil, zero value otherwise.

### GetS3FileSystemImageOk

`func (o *ClusterReportedConfig) GetS3FileSystemImageOk() (*string, bool)`

GetS3FileSystemImageOk returns a tuple with the S3FileSystemImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3FileSystemImage

`func (o *ClusterReportedConfig) SetS3FileSystemImage(v string)`

SetS3FileSystemImage sets S3FileSystemImage field to given value.

### HasS3FileSystemImage

`func (o *ClusterReportedConfig) HasS3FileSystemImage() bool`

HasS3FileSystemImage returns a boolean if a field has been set.

### SetS3FileSystemImageNil

`func (o *ClusterReportedConfig) SetS3FileSystemImageNil(b bool)`

 SetS3FileSystemImageNil sets the value for S3FileSystemImage to be an explicit nil

### UnsetS3FileSystemImage
`func (o *ClusterReportedConfig) UnsetS3FileSystemImage()`

UnsetS3FileSystemImage ensures that no value is present for S3FileSystemImage, not even an explicit nil
### GetS3FileSystemResources

`func (o *ClusterReportedConfig) GetS3FileSystemResources() ClusterReportedResourceRequirements`

GetS3FileSystemResources returns the S3FileSystemResources field if non-nil, zero value otherwise.

### GetS3FileSystemResourcesOk

`func (o *ClusterReportedConfig) GetS3FileSystemResourcesOk() (*ClusterReportedResourceRequirements, bool)`

GetS3FileSystemResourcesOk returns a tuple with the S3FileSystemResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3FileSystemResources

`func (o *ClusterReportedConfig) SetS3FileSystemResources(v ClusterReportedResourceRequirements)`

SetS3FileSystemResources sets S3FileSystemResources field to given value.

### HasS3FileSystemResources

`func (o *ClusterReportedConfig) HasS3FileSystemResources() bool`

HasS3FileSystemResources returns a boolean if a field has been set.

### SetS3FileSystemResourcesNil

`func (o *ClusterReportedConfig) SetS3FileSystemResourcesNil(b bool)`

 SetS3FileSystemResourcesNil sets the value for S3FileSystemResources to be an explicit nil

### UnsetS3FileSystemResources
`func (o *ClusterReportedConfig) UnsetS3FileSystemResources()`

UnsetS3FileSystemResources ensures that no value is present for S3FileSystemResources, not even an explicit nil
### GetFipsMode

`func (o *ClusterReportedConfig) GetFipsMode() FipsModeEnum`

GetFipsMode returns the FipsMode field if non-nil, zero value otherwise.

### GetFipsModeOk

`func (o *ClusterReportedConfig) GetFipsModeOk() (*FipsModeEnum, bool)`

GetFipsModeOk returns a tuple with the FipsMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFipsMode

`func (o *ClusterReportedConfig) SetFipsMode(v FipsModeEnum)`

SetFipsMode sets FipsMode field to given value.

### HasFipsMode

`func (o *ClusterReportedConfig) HasFipsMode() bool`

HasFipsMode returns a boolean if a field has been set.

### SetFipsModeNil

`func (o *ClusterReportedConfig) SetFipsModeNil(b bool)`

 SetFipsModeNil sets the value for FipsMode to be an explicit nil

### UnsetFipsMode
`func (o *ClusterReportedConfig) UnsetFipsMode()`

UnsetFipsMode ensures that no value is present for FipsMode, not even an explicit nil
### GetRunaiOauth2ProxyImage

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyImage() string`

GetRunaiOauth2ProxyImage returns the RunaiOauth2ProxyImage field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxyImageOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyImageOk() (*string, bool)`

GetRunaiOauth2ProxyImageOk returns a tuple with the RunaiOauth2ProxyImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxyImage

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyImage(v string)`

SetRunaiOauth2ProxyImage sets RunaiOauth2ProxyImage field to given value.

### HasRunaiOauth2ProxyImage

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxyImage() bool`

HasRunaiOauth2ProxyImage returns a boolean if a field has been set.

### SetRunaiOauth2ProxyImageNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyImageNil(b bool)`

 SetRunaiOauth2ProxyImageNil sets the value for RunaiOauth2ProxyImage to be an explicit nil

### UnsetRunaiOauth2ProxyImage
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxyImage()`

UnsetRunaiOauth2ProxyImage ensures that no value is present for RunaiOauth2ProxyImage, not even an explicit nil
### GetRunaiOauth2ProxyResources

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyResources() ClusterReportedResourceRequirements`

GetRunaiOauth2ProxyResources returns the RunaiOauth2ProxyResources field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxyResourcesOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyResourcesOk() (*ClusterReportedResourceRequirements, bool)`

GetRunaiOauth2ProxyResourcesOk returns a tuple with the RunaiOauth2ProxyResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxyResources

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyResources(v ClusterReportedResourceRequirements)`

SetRunaiOauth2ProxyResources sets RunaiOauth2ProxyResources field to given value.

### HasRunaiOauth2ProxyResources

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxyResources() bool`

HasRunaiOauth2ProxyResources returns a boolean if a field has been set.

### SetRunaiOauth2ProxyResourcesNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyResourcesNil(b bool)`

 SetRunaiOauth2ProxyResourcesNil sets the value for RunaiOauth2ProxyResources to be an explicit nil

### UnsetRunaiOauth2ProxyResources
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxyResources()`

UnsetRunaiOauth2ProxyResources ensures that no value is present for RunaiOauth2ProxyResources, not even an explicit nil
### GetRunaiOauth2ProxyUsernameLoaderImage

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyUsernameLoaderImage() string`

GetRunaiOauth2ProxyUsernameLoaderImage returns the RunaiOauth2ProxyUsernameLoaderImage field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxyUsernameLoaderImageOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyUsernameLoaderImageOk() (*string, bool)`

GetRunaiOauth2ProxyUsernameLoaderImageOk returns a tuple with the RunaiOauth2ProxyUsernameLoaderImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxyUsernameLoaderImage

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyUsernameLoaderImage(v string)`

SetRunaiOauth2ProxyUsernameLoaderImage sets RunaiOauth2ProxyUsernameLoaderImage field to given value.

### HasRunaiOauth2ProxyUsernameLoaderImage

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxyUsernameLoaderImage() bool`

HasRunaiOauth2ProxyUsernameLoaderImage returns a boolean if a field has been set.

### SetRunaiOauth2ProxyUsernameLoaderImageNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyUsernameLoaderImageNil(b bool)`

 SetRunaiOauth2ProxyUsernameLoaderImageNil sets the value for RunaiOauth2ProxyUsernameLoaderImage to be an explicit nil

### UnsetRunaiOauth2ProxyUsernameLoaderImage
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxyUsernameLoaderImage()`

UnsetRunaiOauth2ProxyUsernameLoaderImage ensures that no value is present for RunaiOauth2ProxyUsernameLoaderImage, not even an explicit nil
### GetUsernameLoaderResources

`func (o *ClusterReportedConfig) GetUsernameLoaderResources() ClusterReportedResourceRequirements`

GetUsernameLoaderResources returns the UsernameLoaderResources field if non-nil, zero value otherwise.

### GetUsernameLoaderResourcesOk

`func (o *ClusterReportedConfig) GetUsernameLoaderResourcesOk() (*ClusterReportedResourceRequirements, bool)`

GetUsernameLoaderResourcesOk returns a tuple with the UsernameLoaderResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsernameLoaderResources

`func (o *ClusterReportedConfig) SetUsernameLoaderResources(v ClusterReportedResourceRequirements)`

SetUsernameLoaderResources sets UsernameLoaderResources field to given value.

### HasUsernameLoaderResources

`func (o *ClusterReportedConfig) HasUsernameLoaderResources() bool`

HasUsernameLoaderResources returns a boolean if a field has been set.

### SetUsernameLoaderResourcesNil

`func (o *ClusterReportedConfig) SetUsernameLoaderResourcesNil(b bool)`

 SetUsernameLoaderResourcesNil sets the value for UsernameLoaderResources to be an explicit nil

### UnsetUsernameLoaderResources
`func (o *ClusterReportedConfig) UnsetUsernameLoaderResources()`

UnsetUsernameLoaderResources ensures that no value is present for UsernameLoaderResources, not even an explicit nil
### GetRunaiOauth2ProxyRedisFQDN

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyRedisFQDN() string`

GetRunaiOauth2ProxyRedisFQDN returns the RunaiOauth2ProxyRedisFQDN field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxyRedisFQDNOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyRedisFQDNOk() (*string, bool)`

GetRunaiOauth2ProxyRedisFQDNOk returns a tuple with the RunaiOauth2ProxyRedisFQDN field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxyRedisFQDN

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyRedisFQDN(v string)`

SetRunaiOauth2ProxyRedisFQDN sets RunaiOauth2ProxyRedisFQDN field to given value.

### HasRunaiOauth2ProxyRedisFQDN

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxyRedisFQDN() bool`

HasRunaiOauth2ProxyRedisFQDN returns a boolean if a field has been set.

### SetRunaiOauth2ProxyRedisFQDNNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyRedisFQDNNil(b bool)`

 SetRunaiOauth2ProxyRedisFQDNNil sets the value for RunaiOauth2ProxyRedisFQDN to be an explicit nil

### UnsetRunaiOauth2ProxyRedisFQDN
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxyRedisFQDN()`

UnsetRunaiOauth2ProxyRedisFQDN ensures that no value is present for RunaiOauth2ProxyRedisFQDN, not even an explicit nil
### GetRunaiOauth2ProxyWorkloadAuthSecretName

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyWorkloadAuthSecretName() string`

GetRunaiOauth2ProxyWorkloadAuthSecretName returns the RunaiOauth2ProxyWorkloadAuthSecretName field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxyWorkloadAuthSecretNameOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyWorkloadAuthSecretNameOk() (*string, bool)`

GetRunaiOauth2ProxyWorkloadAuthSecretNameOk returns a tuple with the RunaiOauth2ProxyWorkloadAuthSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxyWorkloadAuthSecretName

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyWorkloadAuthSecretName(v string)`

SetRunaiOauth2ProxyWorkloadAuthSecretName sets RunaiOauth2ProxyWorkloadAuthSecretName field to given value.

### HasRunaiOauth2ProxyWorkloadAuthSecretName

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxyWorkloadAuthSecretName() bool`

HasRunaiOauth2ProxyWorkloadAuthSecretName returns a boolean if a field has been set.

### SetRunaiOauth2ProxyWorkloadAuthSecretNameNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyWorkloadAuthSecretNameNil(b bool)`

 SetRunaiOauth2ProxyWorkloadAuthSecretNameNil sets the value for RunaiOauth2ProxyWorkloadAuthSecretName to be an explicit nil

### UnsetRunaiOauth2ProxyWorkloadAuthSecretName
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxyWorkloadAuthSecretName()`

UnsetRunaiOauth2ProxyWorkloadAuthSecretName ensures that no value is present for RunaiOauth2ProxyWorkloadAuthSecretName, not even an explicit nil
### GetRunaiOauth2ProxySessionStoreType

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxySessionStoreType() string`

GetRunaiOauth2ProxySessionStoreType returns the RunaiOauth2ProxySessionStoreType field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxySessionStoreTypeOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxySessionStoreTypeOk() (*string, bool)`

GetRunaiOauth2ProxySessionStoreTypeOk returns a tuple with the RunaiOauth2ProxySessionStoreType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxySessionStoreType

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxySessionStoreType(v string)`

SetRunaiOauth2ProxySessionStoreType sets RunaiOauth2ProxySessionStoreType field to given value.

### HasRunaiOauth2ProxySessionStoreType

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxySessionStoreType() bool`

HasRunaiOauth2ProxySessionStoreType returns a boolean if a field has been set.

### SetRunaiOauth2ProxySessionStoreTypeNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxySessionStoreTypeNil(b bool)`

 SetRunaiOauth2ProxySessionStoreTypeNil sets the value for RunaiOauth2ProxySessionStoreType to be an explicit nil

### UnsetRunaiOauth2ProxySessionStoreType
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxySessionStoreType()`

UnsetRunaiOauth2ProxySessionStoreType ensures that no value is present for RunaiOauth2ProxySessionStoreType, not even an explicit nil
### GetRunaiOauth2ProxyOidcIssuerUrl

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyOidcIssuerUrl() string`

GetRunaiOauth2ProxyOidcIssuerUrl returns the RunaiOauth2ProxyOidcIssuerUrl field if non-nil, zero value otherwise.

### GetRunaiOauth2ProxyOidcIssuerUrlOk

`func (o *ClusterReportedConfig) GetRunaiOauth2ProxyOidcIssuerUrlOk() (*string, bool)`

GetRunaiOauth2ProxyOidcIssuerUrlOk returns a tuple with the RunaiOauth2ProxyOidcIssuerUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunaiOauth2ProxyOidcIssuerUrl

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyOidcIssuerUrl(v string)`

SetRunaiOauth2ProxyOidcIssuerUrl sets RunaiOauth2ProxyOidcIssuerUrl field to given value.

### HasRunaiOauth2ProxyOidcIssuerUrl

`func (o *ClusterReportedConfig) HasRunaiOauth2ProxyOidcIssuerUrl() bool`

HasRunaiOauth2ProxyOidcIssuerUrl returns a boolean if a field has been set.

### SetRunaiOauth2ProxyOidcIssuerUrlNil

`func (o *ClusterReportedConfig) SetRunaiOauth2ProxyOidcIssuerUrlNil(b bool)`

 SetRunaiOauth2ProxyOidcIssuerUrlNil sets the value for RunaiOauth2ProxyOidcIssuerUrl to be an explicit nil

### UnsetRunaiOauth2ProxyOidcIssuerUrl
`func (o *ClusterReportedConfig) UnsetRunaiOauth2ProxyOidcIssuerUrl()`

UnsetRunaiOauth2ProxyOidcIssuerUrl ensures that no value is present for RunaiOauth2ProxyOidcIssuerUrl, not even an explicit nil
### GetCustomCAGit

`func (o *ClusterReportedConfig) GetCustomCAGit() ClusterReportedCustomCA`

GetCustomCAGit returns the CustomCAGit field if non-nil, zero value otherwise.

### GetCustomCAGitOk

`func (o *ClusterReportedConfig) GetCustomCAGitOk() (*ClusterReportedCustomCA, bool)`

GetCustomCAGitOk returns a tuple with the CustomCAGit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomCAGit

`func (o *ClusterReportedConfig) SetCustomCAGit(v ClusterReportedCustomCA)`

SetCustomCAGit sets CustomCAGit field to given value.

### HasCustomCAGit

`func (o *ClusterReportedConfig) HasCustomCAGit() bool`

HasCustomCAGit returns a boolean if a field has been set.

### SetCustomCAGitNil

`func (o *ClusterReportedConfig) SetCustomCAGitNil(b bool)`

 SetCustomCAGitNil sets the value for CustomCAGit to be an explicit nil

### UnsetCustomCAGit
`func (o *ClusterReportedConfig) UnsetCustomCAGit()`

UnsetCustomCAGit ensures that no value is present for CustomCAGit, not even an explicit nil
### GetCustomCAS3

`func (o *ClusterReportedConfig) GetCustomCAS3() ClusterReportedCustomCA`

GetCustomCAS3 returns the CustomCAS3 field if non-nil, zero value otherwise.

### GetCustomCAS3Ok

`func (o *ClusterReportedConfig) GetCustomCAS3Ok() (*ClusterReportedCustomCA, bool)`

GetCustomCAS3Ok returns a tuple with the CustomCAS3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomCAS3

`func (o *ClusterReportedConfig) SetCustomCAS3(v ClusterReportedCustomCA)`

SetCustomCAS3 sets CustomCAS3 field to given value.

### HasCustomCAS3

`func (o *ClusterReportedConfig) HasCustomCAS3() bool`

HasCustomCAS3 returns a boolean if a field has been set.

### SetCustomCAS3Nil

`func (o *ClusterReportedConfig) SetCustomCAS3Nil(b bool)`

 SetCustomCAS3Nil sets the value for CustomCAS3 to be an explicit nil

### UnsetCustomCAS3
`func (o *ClusterReportedConfig) UnsetCustomCAS3()`

UnsetCustomCAS3 ensures that no value is present for CustomCAS3, not even an explicit nil
### GetCustomCAOauth2Proxy

`func (o *ClusterReportedConfig) GetCustomCAOauth2Proxy() ClusterReportedCustomCA`

GetCustomCAOauth2Proxy returns the CustomCAOauth2Proxy field if non-nil, zero value otherwise.

### GetCustomCAOauth2ProxyOk

`func (o *ClusterReportedConfig) GetCustomCAOauth2ProxyOk() (*ClusterReportedCustomCA, bool)`

GetCustomCAOauth2ProxyOk returns a tuple with the CustomCAOauth2Proxy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomCAOauth2Proxy

`func (o *ClusterReportedConfig) SetCustomCAOauth2Proxy(v ClusterReportedCustomCA)`

SetCustomCAOauth2Proxy sets CustomCAOauth2Proxy field to given value.

### HasCustomCAOauth2Proxy

`func (o *ClusterReportedConfig) HasCustomCAOauth2Proxy() bool`

HasCustomCAOauth2Proxy returns a boolean if a field has been set.

### SetCustomCAOauth2ProxyNil

`func (o *ClusterReportedConfig) SetCustomCAOauth2ProxyNil(b bool)`

 SetCustomCAOauth2ProxyNil sets the value for CustomCAOauth2Proxy to be an explicit nil

### UnsetCustomCAOauth2Proxy
`func (o *ClusterReportedConfig) UnsetCustomCAOauth2Proxy()`

UnsetCustomCAOauth2Proxy ensures that no value is present for CustomCAOauth2Proxy, not even an explicit nil
### GetExternalAuthUrlEnabled

`func (o *ClusterReportedConfig) GetExternalAuthUrlEnabled() bool`

GetExternalAuthUrlEnabled returns the ExternalAuthUrlEnabled field if non-nil, zero value otherwise.

### GetExternalAuthUrlEnabledOk

`func (o *ClusterReportedConfig) GetExternalAuthUrlEnabledOk() (*bool, bool)`

GetExternalAuthUrlEnabledOk returns a tuple with the ExternalAuthUrlEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalAuthUrlEnabled

`func (o *ClusterReportedConfig) SetExternalAuthUrlEnabled(v bool)`

SetExternalAuthUrlEnabled sets ExternalAuthUrlEnabled field to given value.

### HasExternalAuthUrlEnabled

`func (o *ClusterReportedConfig) HasExternalAuthUrlEnabled() bool`

HasExternalAuthUrlEnabled returns a boolean if a field has been set.

### SetExternalAuthUrlEnabledNil

`func (o *ClusterReportedConfig) SetExternalAuthUrlEnabledNil(b bool)`

 SetExternalAuthUrlEnabledNil sets the value for ExternalAuthUrlEnabled to be an explicit nil

### UnsetExternalAuthUrlEnabled
`func (o *ClusterReportedConfig) UnsetExternalAuthUrlEnabled()`

UnsetExternalAuthUrlEnabled ensures that no value is present for ExternalAuthUrlEnabled, not even an explicit nil
### GetOpenshiftRouteTerminationPolicy

`func (o *ClusterReportedConfig) GetOpenshiftRouteTerminationPolicy() OpenshiftRouteTerminationPolicy`

GetOpenshiftRouteTerminationPolicy returns the OpenshiftRouteTerminationPolicy field if non-nil, zero value otherwise.

### GetOpenshiftRouteTerminationPolicyOk

`func (o *ClusterReportedConfig) GetOpenshiftRouteTerminationPolicyOk() (*OpenshiftRouteTerminationPolicy, bool)`

GetOpenshiftRouteTerminationPolicyOk returns a tuple with the OpenshiftRouteTerminationPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOpenshiftRouteTerminationPolicy

`func (o *ClusterReportedConfig) SetOpenshiftRouteTerminationPolicy(v OpenshiftRouteTerminationPolicy)`

SetOpenshiftRouteTerminationPolicy sets OpenshiftRouteTerminationPolicy field to given value.

### HasOpenshiftRouteTerminationPolicy

`func (o *ClusterReportedConfig) HasOpenshiftRouteTerminationPolicy() bool`

HasOpenshiftRouteTerminationPolicy returns a boolean if a field has been set.

### SetOpenshiftRouteTerminationPolicyNil

`func (o *ClusterReportedConfig) SetOpenshiftRouteTerminationPolicyNil(b bool)`

 SetOpenshiftRouteTerminationPolicyNil sets the value for OpenshiftRouteTerminationPolicy to be an explicit nil

### UnsetOpenshiftRouteTerminationPolicy
`func (o *ClusterReportedConfig) UnsetOpenshiftRouteTerminationPolicy()`

UnsetOpenshiftRouteTerminationPolicy ensures that no value is present for OpenshiftRouteTerminationPolicy, not even an explicit nil
### GetDistGPUNetworkAcceleration

`func (o *ClusterReportedConfig) GetDistGPUNetworkAcceleration() bool`

GetDistGPUNetworkAcceleration returns the DistGPUNetworkAcceleration field if non-nil, zero value otherwise.

### GetDistGPUNetworkAccelerationOk

`func (o *ClusterReportedConfig) GetDistGPUNetworkAccelerationOk() (*bool, bool)`

GetDistGPUNetworkAccelerationOk returns a tuple with the DistGPUNetworkAcceleration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistGPUNetworkAcceleration

`func (o *ClusterReportedConfig) SetDistGPUNetworkAcceleration(v bool)`

SetDistGPUNetworkAcceleration sets DistGPUNetworkAcceleration field to given value.

### HasDistGPUNetworkAcceleration

`func (o *ClusterReportedConfig) HasDistGPUNetworkAcceleration() bool`

HasDistGPUNetworkAcceleration returns a boolean if a field has been set.

### SetDistGPUNetworkAccelerationNil

`func (o *ClusterReportedConfig) SetDistGPUNetworkAccelerationNil(b bool)`

 SetDistGPUNetworkAccelerationNil sets the value for DistGPUNetworkAcceleration to be an explicit nil

### UnsetDistGPUNetworkAcceleration
`func (o *ClusterReportedConfig) UnsetDistGPUNetworkAcceleration()`

UnsetDistGPUNetworkAcceleration ensures that no value is present for DistGPUNetworkAcceleration, not even an explicit nil
### GetDinfGPUNetworkAcceleration

`func (o *ClusterReportedConfig) GetDinfGPUNetworkAcceleration() bool`

GetDinfGPUNetworkAcceleration returns the DinfGPUNetworkAcceleration field if non-nil, zero value otherwise.

### GetDinfGPUNetworkAccelerationOk

`func (o *ClusterReportedConfig) GetDinfGPUNetworkAccelerationOk() (*bool, bool)`

GetDinfGPUNetworkAccelerationOk returns a tuple with the DinfGPUNetworkAcceleration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDinfGPUNetworkAcceleration

`func (o *ClusterReportedConfig) SetDinfGPUNetworkAcceleration(v bool)`

SetDinfGPUNetworkAcceleration sets DinfGPUNetworkAcceleration field to given value.

### HasDinfGPUNetworkAcceleration

`func (o *ClusterReportedConfig) HasDinfGPUNetworkAcceleration() bool`

HasDinfGPUNetworkAcceleration returns a boolean if a field has been set.

### SetDinfGPUNetworkAccelerationNil

`func (o *ClusterReportedConfig) SetDinfGPUNetworkAccelerationNil(b bool)`

 SetDinfGPUNetworkAccelerationNil sets the value for DinfGPUNetworkAcceleration to be an explicit nil

### UnsetDinfGPUNetworkAcceleration
`func (o *ClusterReportedConfig) UnsetDinfGPUNetworkAcceleration()`

UnsetDinfGPUNetworkAcceleration ensures that no value is present for DinfGPUNetworkAcceleration, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


