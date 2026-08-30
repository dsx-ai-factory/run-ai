// Step configuration interface

import type { ISettingEmailServer } from "@/models/setting.model";
import { type PermissionCheck } from "@/stores/permissions.store";
import { Action } from "@/swagger-models/authorization-client";
import { ResourceType } from "@/swagger-models/authorization-client";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";

export interface IStepConfig {
  id: number;
  name: string;
  subSteps: number;
  continueLabel?: string;
  getSubSteps?: () => number; // Function to calculate substeps dynamically
}

export const DEFAULT_CLUSTER_VERSION = "2.22.0"; // Default version if fetching fails

// Cluster data interface for consolidated cluster information
export interface ClusterData {
  clusterName: string;
  clusterVersion: string;
  installationOption: string;
  clusterLocation: string;
  clusterUrl: string;
  hasClusterValues: boolean;
  isClusterConnected: boolean;
  clusterId?: string;
  // Installation-related fields
  installationCommand?: string;
  installationError?: string | null;
  installationErrorType?: InstallationErrorType | null;
  isExistingCluster?: boolean;
}

// Type for updatable cluster data keys (excluding read-only properties)
export type ClusterDataKey = keyof Pick<
  ClusterData,
  "clusterName" | "installationOption" | "clusterLocation" | "clusterUrl" | "isClusterConnected"
>;

// Team identification method enum
export enum TeamIdentificationMethod {
  MemberEmails = "member-emails",
  GroupName = "group-name",
}

// Team data interface for consolidated team information
export interface TeamData {
  letSomeoneElseFill: boolean;
  delegateEmail: string; // Email for delegation when letSomeoneElseFill is true
  teamName: string; // Team name when user fills themselves
  memberEmails: string[]; // Team member emails when user fills themselves
  identificationMethod: TeamIdentificationMethod; // How to identify team members
  groupName: string; // Group name for IDP group identification
}

// Type for updatable team data keys
export type TeamDataKey = keyof Pick<
  TeamData,
  "letSomeoneElseFill" | "delegateEmail" | "teamName" | "memberEmails" | "identificationMethod" | "groupName"
>;

// Import API types for SSO data
import type {
  OidcCreationData,
  SamlCreationData,
  Mappers,
  Idp,
  UserPostResponse,
} from "@/swagger-models/identity-manager-client";

// SSO data interface for consolidated SSO information
export interface SsoData {
  protocol: string;
  idpMetadata: string | object; // String for SAML XML, object for OIDC metadata (backward compatibility)
  runaiMetadata: string;
  isDelegated: boolean;
  delegateEmail: string;
  hasIdpCreated: boolean; // Flag to indicate successful IDP creation
  // New structured data for API calls
  oidcData?: OidcCreationData;
  samlData?: SamlCreationData;
  mappers: Mappers;
  idp?: Idp;
  // Error handling fields
  ssoError?: string | null;
  ssoErrorType?: SsoErrorType | null;
}

// Type for updatable SSO data keys
export type SsoDataKey = keyof SsoData;

// Type for SSO data update emit validation function
export type SsoDataUpdateEmitValidator = (
  key: SsoDataKey,
  value: string | boolean | object | Idp | Mappers | OidcCreationData | SamlCreationData | null | undefined,
) => boolean;

// Reusable SSO data update emit validation function
export const ssoDataUpdateEmitValidator: SsoDataUpdateEmitValidator = (
  key: SsoDataKey,
  value: string | boolean | object | Idp | Mappers | OidcCreationData | SamlCreationData | null | undefined,
) => !!key && value !== undefined;

// Provisioning data interface for consolidated provisioning information
export interface ProvisioningData {
  completionState?: IOnboardingCompletionState;
  failedSubjectIds: string[];
  createdUsers?: UserPostResponse[];
}

// Type for updatable provisioning data keys
export type ProvisioningDataKey = keyof ProvisioningData;

export interface EmailServerData {
  settings: ISettingEmailServer;
  isConfigured: boolean;
  saveError: boolean;
}

export type EmailServerDataKey = keyof Pick<EmailServerData, "settings">;

// Onboarding state interface
export interface OnboardingState {
  currentStep: number;
  currentSubStep: number;
  completedSteps: number[];
  visitedSubSteps: Record<number, Set<number>>;
  canContinue: boolean;
  isLoading: boolean;
  clusterData: ClusterData;
  ssoData: SsoData;
  teamData: TeamData;
  provisioningData: ProvisioningData;
  emailServerData: EmailServerData;
}

//Onboarding completion state
export enum IOnboardingCompletionState {
  Completed = "Completed",
  CompletedWithoutProject = "CompletedWithoutProject",
  ProjectCreationFailed = "ProjectCreationFailed",
  AccessRulesCreationFailed = "AccessRulesCreationFailed",
  AdminUserCreationFailed = "AdminUserCreationFailed",
}

export const onboardingCompletionStateSvg: Record<IOnboardingCompletionState, string> = {
  [IOnboardingCompletionState.Completed]: "lets-do-this",
  [IOnboardingCompletionState.CompletedWithoutProject]: "beep-boop",
  [IOnboardingCompletionState.ProjectCreationFailed]: "all-by-myself",
  [IOnboardingCompletionState.AccessRulesCreationFailed]: "lunchtime",
  [IOnboardingCompletionState.AdminUserCreationFailed]: "up-and-away",
};

// Onboarding cluster installation
export interface InstallationData {
  createdCluster: DisplayedCluster | null;
  clientSecret: string;
  installationInstructions: string;
  isExistingCluster: boolean;
  installationCommand: string;
}

export enum InstallationErrorType {
  ClusterCreation = "cluster-creation",
  InstallationInstructions = "installation-instructions",
}
export interface InstallationError {
  type: InstallationErrorType;
  message: string;
}

export enum SsoErrorType {
  SamlMetadataRetrieval = "saml-metadata-retrieval",
  OidcMetadataRetrieval = "oidc-metadata-retrieval",
  IdpCreation = "idp-creation",
  MissingData = "missing-data",
}

export interface InstallationResult {
  success: boolean;
  data?: InstallationData;
  error?: InstallationError;
}

// Onboarding Step Enum
export enum OnboardingStep {
  Landing = 0,
  ClusterSetup = 1,
  SsoSetup = 2,
  EmailServerSetup = 3,
  TeamSetup = 4,
  Completion = 5,
}

// SSO Substep Enum
export enum SsoSubstep {
  Protocol = 1,
  IdpMetadata = 2,
  RunaiMetadata = 3,
}

export enum TeamSetupSubstep {
  CreateTeam = 1,
  TeamCredentials = 2,
}

export type StepRequiredPermissions = Partial<Record<OnboardingStep, PermissionCheck[]>>;

export const stepsRequiredPermissions: StepRequiredPermissions = {
  [OnboardingStep.ClusterSetup]: [{ resourceType: ResourceType.Cluster, action: Action.Create }],
  [OnboardingStep.SsoSetup]: [
    { resourceType: ResourceType.SecuritySettings, action: Action.Create },
    { resourceType: ResourceType.SecuritySettings, action: Action.Update },
    { resourceType: ResourceType.Users, action: Action.Create },
    { resourceType: ResourceType.AccessRules, action: Action.Create },
  ],
  [OnboardingStep.TeamSetup]: [
    { resourceType: ResourceType.Project, action: Action.Create },
    { resourceType: ResourceType.AccessRules, action: Action.Create },
    { resourceType: ResourceType.Users, action: Action.Create },
  ],
  [OnboardingStep.EmailServerSetup]: [
    { resourceType: ResourceType.Settings, action: Action.Create },
    { resourceType: ResourceType.Settings, action: Action.Update },
  ],
};
