import type { EmailAuthType } from "@/swagger-models/notifications-proxy-client";

export enum ENotificationType {
  Email = "email",
}

export interface SmtpConfig {
  type: ENotificationType;
  config: Config;
}

export interface Config {
  user: string;
  password: string;
  smtp_tls_enabled: boolean;
  smtp_host: string;
  smtp_port: number;
  auth_type: EmailAuthType;
  recipients?: string[];
  from_display_name: string;
  from: string;
  direct_notifications: boolean;
  subject: string;
}

export interface EnableSmtpConfig {
  route_config: RouteConfig;
  route_id: string;
  tenant_id: string;
}

export interface VerificationObject {
  errorCode?: number;
  errorMessage?: string;
}

export interface RouteConfig {
  listener_id: string;
  notify_ids: [string];
  filtering: {
    policy: string;
  };
}
