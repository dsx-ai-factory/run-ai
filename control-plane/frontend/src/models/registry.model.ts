export type RegistrySpec = {
  password?: string;
  credentialKind: string;
  user: string;
  url: string;
};

export type AssetMeta = {
  name: string;
  scope?: string;
};

export interface Registry {
  meta: AssetMeta;
  spec: RegistrySpec;
}

export interface AssetResponse extends AssetMeta {
  id: string;
}

export interface RegistryResponse extends Registry {
  meta: AssetResponse;
}
export interface RegistryResponse {
  entries: Array<RegistryResponse>;
}

export interface RepositoriesListResponse {
  repositories: string[];
}

export interface TagsListResponse {
  tags: string[];
}
