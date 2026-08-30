export interface ILegacyRoleTransitionRow {
  id: string;
  legacyRole: string;
  roleId: number;
  accessRuleCount: number;
  newRole: string | undefined;
}
