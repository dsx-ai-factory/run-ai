import type { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

export enum ETokenProvisionMethods {
  Manually = "manually",
  FromSecret = "from_secret",
}

export interface IHuggingFaceAccessSectionModel {
  modelName: string;
  token?: string;
  selectedCredentials?: string;
  credentialsSource?: ERunaiCredentialSource;
  selectedTokenProvisionMethod: ETokenProvisionMethods;
}

export enum EHuggingFaceAccessSectionEvents {
  UpdateModelValue = "update:model-value",
}

export interface IHuggingFaceModelAndTokenValidationRes {
  modelName: string;
  hfToken?: string;
  gated?: boolean;
  authorized?: boolean;
  language?: boolean;
  errorMessage?: string;
}
