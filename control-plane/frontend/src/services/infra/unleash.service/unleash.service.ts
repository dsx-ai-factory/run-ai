import { unleashConfigService } from "./unleash-config.service";
import { EFeatureToggle } from "@/services/infra/unleash.service/unleash.model";

export const unleashService = {
  isUpdateAssetEnabled,
  isFeatureEnabled,
  showNgcIntegration,
  showSubmitAIApplication,
};

function isFeatureEnabled(featureToggle: EFeatureToggle): boolean {
  return unleashConfigService.isFeatureEnabled(featureToggle);
}

function isUpdateAssetEnabled(): boolean {
  return unleashConfigService.isFeatureEnabled(EFeatureToggle.ShowUpdateAssetEnabled);
}

function showNgcIntegration(): boolean {
  return isFeatureEnabled(EFeatureToggle.ShowNgcIntegration);
}

function showSubmitAIApplication(): boolean {
  return isFeatureEnabled(EFeatureToggle.ShowSubmitAIApplication);
}
