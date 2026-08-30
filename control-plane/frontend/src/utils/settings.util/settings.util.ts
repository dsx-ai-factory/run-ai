// models
import type { IIdpMapperItem, ISettingEmailServer } from "@/models/setting.model";
import { MappersKeys, DEFAULT_SMTP_PORT } from "@/models/setting.model";
import type {
  IdpCreationRequest,
  Mappers,
  OcpCreationData,
  OidcCreationData,
  SamlCreationData,
  Idp,
} from "@/swagger-models/identity-manager-client";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import type { EmailChannelConfig } from "@/swagger-models/notifications-proxy-client";

// utils
import { alertUtil } from "@/utils/alert.util";
import { stringUtil } from "@/utils/string.util";

// store
import type { Setting } from "@/models/setting.model";
import { EmailAuthType } from "@/swagger-models/notifications-proxy-client";
import { docsUtil } from "@/utils/docs.util/docs.util";
import type { QNotifyCreateOptions } from "quasar";
import DOCS from "@/constants/docs";

export const settingsUtil = {
  getIdpMappersItems,
  isIdpMappersEqual,
  getMappersFromIdpMapperItems,
  createIdpCreationRequest,
  getSettingErrorMessage,
  getToggleSectionCollapsedTitle,
  isActive,
  emailSettingsToSmtpConfig,
  smtpConfigToEmailSettings,
  getDefaultEntityId,
  getDefaultIdpConfig,
};

/** True when the setting's value is a boolean or the string "true" or "false" (interpretable as on/off). */
export function isActive(value: unknown): value is Setting {
  return (
    value != null &&
    (typeof (value as Setting)?.value === "boolean" ||
      (value as Setting)?.value === "true" ||
      (value as Setting)?.value === "false")
  );
}

function getIdpMappersItems(mappers: Mappers): Record<string, IIdpMapperItem> {
  return {
    uid: {
      key: MappersKeys.Uid,
      label: "Linux user ID",
      value: mappers?.uid || "",
    },
    gid: {
      key: MappersKeys.Gid,
      label: "Linux group ID",
      value: mappers?.gid || "",
    },
    groups: {
      key: MappersKeys.Groups,
      label: "User role groups",
      value: mappers?.groups || "",
    },
    supplementaryGroups: {
      key: MappersKeys.SupplementaryGroups,
      label: "SUPPLEMENTARYGROUPS",
      value: mappers?.supplementaryGroups || "",
    },
    email: {
      key: MappersKeys.Email,
      label: "Email",
      value: mappers?.email || "",
    },
    firstName: {
      key: MappersKeys.FirstName,
      label: "First name",
      value: mappers?.firstName || "",
    },
    lastName: {
      key: MappersKeys.LastName,
      label: "Last name",
      value: mappers?.lastName || "",
    },
  };
}

function isIdpMappersEqual(mappers: Mappers, items: Record<string, IIdpMapperItem>): boolean {
  return (
    mappers.uid === items[MappersKeys.Uid]?.value &&
    mappers.gid === items[MappersKeys.Gid]?.value &&
    mappers.groups === items[MappersKeys.Groups]?.value &&
    mappers.supplementaryGroups === items[MappersKeys.SupplementaryGroups]?.value &&
    mappers.email === items[MappersKeys.Email]?.value &&
    mappers.firstName === items[MappersKeys.FirstName]?.value &&
    mappers.lastName === items[MappersKeys.LastName]?.value
  );
}

function getMappersFromIdpMapperItems(items: Record<string, IIdpMapperItem>): Mappers {
  return {
    uid: items[MappersKeys.Uid]?.value,
    gid: items[MappersKeys.Gid]?.value,
    groups: items[MappersKeys.Groups]?.value,
    supplementaryGroups: items[MappersKeys.SupplementaryGroups]?.value,
    email: items[MappersKeys.Email]?.value,
    firstName: items[MappersKeys.FirstName]?.value,
    lastName: items[MappersKeys.LastName]?.value,
  };
}

function createIdpCreationRequest(
  alias: string,
  idpNewFields: OidcCreationData | SamlCreationData | OcpCreationData,
  mappers: Mappers,
): IdpCreationRequest | null {
  switch (alias) {
    case IdpTypeEnum.Oidc:
      return {
        type: alias,
        oidcData: idpNewFields as OidcCreationData,
        mappers: mappers,
      };

    case IdpTypeEnum.Saml:
      return {
        type: alias,
        samlData: idpNewFields as SamlCreationData,
        mappers: mappers,
      };

    case IdpTypeEnum.OpenshiftV4:
      return {
        type: alias,
        ocpData: idpNewFields as OcpCreationData,
        mappers: mappers,
      };

    default:
      return null;
  }
}

function getSettingErrorMessage(): QNotifyCreateOptions {
  const message = DOCS.SUPPORT.URL
    ? `Settings couldn't be saved. Try again. If the issue persists, ${docsUtil.getSupportAnchor("contact Run:ai support")}.`
    : "Settings couldn't be saved. Try again.";
  return alertUtil.getError(message, { html: true });
}

function getToggleSectionCollapsedTitle(settings: object): string {
  let enabled = 0;
  let disabled = 0;
  Object.values(settings)
    .filter(isActive)
    .forEach((value) => {
      if (!value) return;
      if (stringUtil.toBoolean(value.value)) {
        enabled++;
      } else {
        disabled++;
      }
    });
  return `Enabled (${enabled}) / Disabled (${disabled})`;
}

function emailSettingsToSmtpConfig(emailSettings: ISettingEmailServer): EmailChannelConfig {
  const baseConfig = {
    smtpHost: emailSettings.smtpHost,
    from: emailSettings.fromEmail,
    smtpPort: emailSettings.smtpPort,
    authType: emailSettings.authType,
    smtpTlsEnabled: emailSettings.authType !== EmailAuthType.AuthNone,
    recipients: [],
    fromDisplayName: emailSettings.fromName,
    directNotifications: true,
    subject: "{{ .title }}",
  };
  if (emailSettings.authType === EmailAuthType.AuthNone) {
    return baseConfig;
  }
  return {
    ...baseConfig,
    username: emailSettings.username,
    password: emailSettings.password,
  };
}

function smtpConfigToEmailSettings(emailChannelConfig: EmailChannelConfig): ISettingEmailServer {
  return {
    smtpHost: emailChannelConfig.smtpHost || "",
    smtpPort: emailChannelConfig.smtpPort || DEFAULT_SMTP_PORT,
    username: emailChannelConfig.username || "",
    password: emailChannelConfig.password || "",
    fromEmail: emailChannelConfig.from || "",
    fromName: emailChannelConfig.fromDisplayName || "",
    authType: emailChannelConfig.authType || EmailAuthType.AuthNone,
  };
}

function getDefaultEntityId(hostName: string): string {
  return `runai-${hostName}/runai`;
}

function getDefaultIdpConfig(idpType: IdpTypeEnum, authHost: string): Idp | undefined {
  if (!idpType) return undefined;
  return {
    alias: idpType,
    type: idpType,
    redirectUri: getDefaultRedirectUri(authHost),
  };
}

function getDefaultRedirectUri(authHost: string): string {
  return `${authHost}/broker/<idp-name>/endpoint`;
}
