import type { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import type { ETokenProvisionMethods } from "../hugging-face-access-section/hugging-face-access-section.model";

export interface IHuggingFaceModelAndAccessModel {
  modelName: string;
  token?: string;
  selectedCredentials?: string;
  credentialsSource?: ERunaiCredentialSource;
  selectedTokenProvisionMethod: ETokenProvisionMethods;
  accessFromStorage?: boolean;
}
