export enum BACKEND_MESSAGES {
  MODEL_WAS_NOT_FOUND_ON_HUGGING_FACE = "the provided model was not found on hugging face",
}

export interface IHuggingFaceModelValidationRes {
  modelName: string;
  gated?: boolean;
  authorized?: boolean;
  language?: boolean;
  errorMessage?: string;
  warningMessage?: string;
}

export enum EHuggingFaceModelEvents {
  UpdateModelValue = "update:model-value",
  IsValid = "is-valid",
}

export interface IHuggingFaceModelSectionEvent {
  valid: boolean;
  gated: boolean;
}
