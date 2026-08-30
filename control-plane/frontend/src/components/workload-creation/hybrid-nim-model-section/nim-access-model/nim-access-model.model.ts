import type {
  ERunaiCredentialSource,
  ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import type { IDockerRegistrySelectorModel } from "./docker-registry-selector/docker-registry-selector.model";

export enum ENimAccessModelMethod {
  Manually = "manually",
  FromSecret = "from_secret",
}

export interface INimAccessModelModel {
  apiKey?: string;
  selectedCredentials?: string;
  accessFromStorage?: boolean;
  credentialsSource?: ERunaiCredentialSource;
  credentialType?: ERunaiCredentialType;
  dockerRegistry?: IDockerRegistrySelectorModel;
}

export enum ENimAccessModelEvents {
  UpdateModelValue = "update:model-value",
}
