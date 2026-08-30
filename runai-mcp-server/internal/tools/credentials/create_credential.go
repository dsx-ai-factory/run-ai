package credentials

import (
	"context"
	"fmt"
	"net/http"

	assets "github.com/run-ai/runai/runai-clients/assets-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

func createCredential(
	ctx context.Context,
	client *runaiclient.Client,
	in manageCredentialsInput,
	out manageCredentialsOutput,
) (manageCredentialsOutput, error) {
	adapter, err := credentialAdapterByType(in.CredentialType)
	if err != nil {
		return manageCredentialsOutput{}, err
	}
	created, resp, err := adapter.create(
		ctx, client, credentialCreationMeta(in), in.Spec)
	if err != nil {
		return manageCredentialsOutput{}, fmt.Errorf(
			"create %s credential %q: %w", in.CredentialType, in.Name, runaimcp.MapHTTPError(resp, err))
	}

	out.CredentialID = created.Id
	out.Name = created.Name
	out.CredentialType = string(created.Kind)
	out.Private = created.GetPrivate()
	out.After = credentialSnapshot(created)
	out.Diff = common.DiffMaps(nil, out.After)
	out.Applied = true
	return out, nil
}

func (accessKeyAdapter) create(
	ctx context.Context,
	client *runaiclient.Client,
	meta assets.AssetCreationRequest,
	in *credentialSpec,
) (assets.AssetMeta, *http.Response, error) {
	spec := assets.AccessKeyCreationSpec{}
	if in.ExistingSecretName != "" {
		spec.SetExistingSecretName(in.ExistingSecretName)
	} else {
		spec.SetAccessKeyId(in.AccessKeyID)
		spec.SetSecretAccessKey(in.SecretAccessKey)
	}

	result, resp, err := client.Assets.CredentialsAPI.CreateAccessKey(ctx).
		AccessKeyCreationRequest(assets.AccessKeyCreationRequest{Meta: meta, Spec: spec}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (passwordAdapter) create(
	ctx context.Context,
	client *runaiclient.Client,
	meta assets.AssetCreationRequest,
	in *credentialSpec,
) (assets.AssetMeta, *http.Response, error) {
	spec := assets.PasswordCreationSpec{}
	if in.ExistingSecretName != "" {
		spec.SetExistingSecretName(in.ExistingSecretName)
	} else {
		spec.SetUser(in.User)
		spec.SetPassword(in.Password)
	}

	result, resp, err := client.Assets.CredentialsAPI.CreatePassword(ctx).
		PasswordCreationRequest(assets.PasswordCreationRequest{Meta: meta, Spec: spec}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (dockerRegistryAdapter) create(
	ctx context.Context,
	client *runaiclient.Client,
	meta assets.AssetCreationRequest,
	in *credentialSpec,
) (assets.AssetMeta, *http.Response, error) {
	spec := assets.DockerRegistryCreationSpec{}
	if in.ExistingSecretName != "" {
		spec.SetExistingSecretName(in.ExistingSecretName)
	} else {
		spec.SetUser(in.User)
		spec.SetPassword(in.Password)
		spec.SetUrl(in.URL)
	}

	result, resp, err := client.Assets.CredentialsAPI.CreateDockerRegistry(ctx).
		DockerRegistryCreationRequest(assets.DockerRegistryCreationRequest{Meta: meta, Spec: spec}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (genericSecretAdapter) create(
	ctx context.Context,
	client *runaiclient.Client,
	meta assets.AssetCreationRequest,
	in *credentialSpec,
) (assets.AssetMeta, *http.Response, error) {
	spec := assets.GenericSecretCreationSpec{}
	if in.ExistingSecretName != "" {
		spec.SetExistingSecretName(in.ExistingSecretName)
	} else {
		pairs := make([]assets.KeyValuePair, 0, len(in.KeyValuePairs))
		for _, pair := range in.KeyValuePairs {
			item := assets.KeyValuePair{}
			item.SetKey(pair.Key)
			item.SetValue(pair.Value)
			pairs = append(pairs, item)
		}
		spec.SetKeyValuePairs(pairs)
	}

	result, resp, err := client.Assets.CredentialsAPI.CreateGenericSecret(ctx).
		GenericSecretCreationRequest(assets.GenericSecretCreationRequest{Meta: meta, Spec: spec}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (ngcAPIKeyAdapter) create(
	ctx context.Context,
	client *runaiclient.Client,
	meta assets.AssetCreationRequest,
	in *credentialSpec,
) (assets.AssetMeta, *http.Response, error) {
	spec := assets.NgcApiKeyCreationSpec{}
	spec.SetNgcApiKey(in.NGCAPIKey)

	result, resp, err := client.Assets.CredentialsAPI.CreateNgcApiKey(ctx).
		NgcApiKeyCreationRequest(assets.NgcApiKeyCreationRequest{Meta: meta, Spec: spec}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func credentialCreationMeta(in manageCredentialsInput) assets.AssetCreationRequest {
	scope := assets.Scope(in.Scope)
	if in.Private {
		scope = assets.SCOPE_TENANT
	}

	meta := assets.AssetCreationRequest{
		Name:  in.Name,
		Scope: scope,
	}
	meta.SetPrivate(in.Private)
	if in.Description != nil {
		meta.SetDescription(*in.Description)
	}
	if in.ClusterID != "" {
		meta.SetClusterId(in.ClusterID)
	}
	if in.DepartmentID != "" {
		meta.SetDepartmentId(in.DepartmentID)
	}
	if in.ProjectID != nil {
		meta.SetProjectId(*in.ProjectID)
	}
	return meta
}
