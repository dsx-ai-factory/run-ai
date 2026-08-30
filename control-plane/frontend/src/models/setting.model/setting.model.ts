import type { Mappers } from "@/swagger-models/identity-manager-client";
import type { Setting as ServerSetting } from "@/swagger-models/temp-backend-settings-client";

export interface Setting
  extends Omit<ServerSetting, "key" | "source" | "category" | "label" | "description" | "type" | "active"> {
  key?: string;
  source?: string;
  category?: string;
  label?: string;
  description?: string;
  type?: string;
  active?: boolean;
}
import { EmailAuthType } from "@/swagger-models/notifications-proxy-client";
import { SettingsKeyEnum } from "@/swagger-models/identity-manager-client";
import DOCS from "@/constants/docs";

export const enum MappersKeys {
  Uid = "uid",
  Gid = "gid",
  Groups = "groups",
  SupplementaryGroups = "supplementaryGroups",
  Email = "email",
  FirstName = "firstName",
  LastName = "lastName",
}

export enum SettingKeys {
  ProjectManualOverQuota = "project.quota.manual_over_quota",
  CPUResourcesQuota = "project.quota.cpu_resources",
  AllowSwap = "tenant.allow_swap",
  AllowGPT2Model = "tenant.allow_gpt2_model",
  AdvancedGPUMetrics = "tenant.enable_advanced_gpu_metrics",
  LimitQuotaOverSubscription = "tenant.limit_quota_over_subscription",
  EnablePolicies = "tenant.enable_policies",
  EnableModelCatalog = "tenant.enable_model_catalog",
  EnableNIM = "tenant.enable_nim",
  AdminMessage = "tenant.admin_message",
  GpuOverProvisioning = "tenant.gpu_over_provisioning",

  SmtpEnabled = "tenant.smtp_enabled",
  DeletedWorkloadsView = "tenant.deleted_workloads_view",
  EnableHybridSubmission = "tenant.enable_hybrid_submission",
  EnableDataVolumes = "tenant.enable_data_volumes",
  AutoRedirectToSSO = "tenant.auto_redirect_to_sso",
  EnableTemplatesV2 = "tenant.enable_templates_v2",
  EnableNgcIntegration = "tenant.enable_ngc_integration",
  YamlSubmission = "yamlSubmission",
  ApplicationSubmission = "applicationSubmission",
}

export const defaultMappers: Mappers = {
  uid: "UID",
  gid: "GID",
  groups: "GROUPS",
  supplementaryGroups: "SUPPLEMENTARYGROUPS",
  email: "email",
  firstName: "firstName",
  lastName: "lastName",
};

export const defaultOcpMappers: Mappers = {
  uid: "metadata.uid",
  gid: "metadata.gid",
  groups: "groups",
  supplementaryGroups: "supplementarygroups",
  email: "metadata.name",
  firstName: "firstName",
  lastName: "lastName",
};

export const settingUpdateSuccessMessage = "Settings saved";
export const defaultEntityId = "runai-app.run.ai/runai";
export const urlHintMessage = "If left empty, the latest metadata file will be used";
const withEmailSupportCopy = (message: string): string =>
  DOCS.SUPPORT.URL ? `${message} If the issue persists, contact Run:ai support.` : message;

export const EmailErrorMessages = {
  EMAIL_SAVED_ERROR: withEmailSupportCopy("Email server couldn't be saved. Try again."),
  DELETE_SERVER_ERROR: withEmailSupportCopy("Email server couldn't be deleted. Try again."),
  EMAIL_NOTIFICATION_ENABLE_ERROR: withEmailSupportCopy("Email notifications couldn't be enabled. Try again."),
  EMAIL_NOTIFICATION_DISABLE_ERROR: withEmailSupportCopy("Email notifications couldn't be disabled. Try again."),
};

export const LogoConstants = {
  TRANSPARENT_PNG_BASE64: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+V4nEAAAAASUVORK5CYII=",
  MAX_SIZE_KB: 128,
  ALLOWED_TYPES: ["image/png", "image/svg+xml"] as Array<string>,
} as const;

export const DEFAULT_SMTP_PORT = 587;
export const DEFAULT_EMPTY_EMAIL_SETTINGS: ISettingEmailServer = {
  smtpHost: "",
  authType: EmailAuthType.AuthPlain,
  smtpPort: DEFAULT_SMTP_PORT,
  username: "",
  password: "",
  fromEmail: "",
  fromName: "",
};
export enum EBadgeLabel {
  Experimental = "Experimental",
  Beta = "Beta",
  Legacy = "Legacy",
  New = "New",
}

export const AlphaStage = "Alpha";

export type TSettingValue = boolean | string | number | IAdminMessage;

export const DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES = 0;

export interface IAdminMessage {
  msg: string;
  id: string;
  createdAt: Date;
  enableDismiss: boolean;
}

export interface IAdminMessageSetting extends Omit<Setting, "value"> {
  value: IAdminMessage;
}

export interface IFlagSetting {
  value: boolean;
  displayName?: string;
  description?: string;
  stage?: string;
}

export interface IEnableNIMSetting extends IFlagSetting {}

export interface IDeletedWorkloadsViewSetting extends IFlagSetting {}

export interface IAllowGPT2ModelSetting extends IFlagSetting {}

export interface IEnableModelCatalogSetting extends IFlagSetting {}

export interface IEnablePoliciesSetting extends IFlagSetting {}

export interface IEnableHybridSubmissionSetting extends IFlagSetting {}

export interface IEnableTemplatesV2Setting extends IFlagSetting {}

export interface IGpuOverProvisioningSetting extends IFlagSetting {}

export interface IEnableNgcIntegrationSetting extends IFlagSetting {}

export interface IEnableDataVolumesSetting extends IFlagSetting {}

export interface IEnableAdvancedGpuMetricsSetting extends IFlagSetting {}

export interface ILimitQuotaOverSubscriptionSetting extends IFlagSetting {}

export interface IProjectManualOverQuotaSetting extends IFlagSetting {}

export interface ICPUResourcesQuotaSetting extends IFlagSetting {}

export interface IAnalyticsSectionSettings {
  advancedGpuMetrics?: Setting;
}

export interface IWorkloadsSectionSettings {
  policies?: Setting;
  models?: Setting;
  allowGPT2Model?: Setting;

  deletedWorkloadsView?: Setting;
  nim?: Setting;
  hybridSubmission?: Setting;
  dataVolumes?: Setting;
  templatesV2?: Setting;
  ngcIntegration?: Setting;
  yamlSubmission?: Setting;
  applicationSubmission?: Setting;
}

export interface IResourceSectionSettings {
  projectsOverDepartmentQuota?: Setting;
  cpuOverQuota?: Setting;
  overQuotaPriority?: Setting;
  gpuOptimization?: Setting;
}

export interface IIdpMapperItem {
  key: string;
  label: string;
  value: string;
}

export interface IWrapperBoxSettingMessages {
  headerTitle: string;
  editToolTip: string;
  removeToolTip: string;
  removeDisable: string;
  confirmRemoveTitle: string;
  confirmRemoveSubTitle: string;
}

export interface ISettingEmailServer {
  smtpHost: string;
  smtpPort: number;
  username: string;
  password: string;
  fromEmail: string;
  fromName: string;
  authType: EmailAuthType;
}

export interface SamlConfigSettings {
  idpEntityId: string;
  singleSignOnServiceUrl: string;
  signingCertificate: string;
}

export const autoRedirectSSOSetting: Setting = {
  label: "Auto-Redirect to SSO",
  value: "false",
  key: SettingsKeyEnum.AutoRedirectSso,
  description: `Enable to automatically redirect users to the configured identity provider’s login page`,
};
