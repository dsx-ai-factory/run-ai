export interface IHybridEnvironmentSectionDisableFields {
  servingPort?: string;
  protocol?: boolean;
  tools?: string;
  toolsAccess?: string;
  security?: {
    uidGidSource?: string;
    capabilities?: string;
  };
}
