export enum EAccessOptions {
  EVERYONE = "everyone",
  GROUPS = "groups",
  SPECIFIC_USERS = "specific-users",
  PRIVATE = "private",
  PUBLIC = "public",
}

export interface IConnectionAccess {
  authorizedUsers?: Array<string> | null;
  authorizedGroups?: Array<string> | null;
  accessOption: EAccessOptions;
}

export enum EToolLabel {
  UserEmailOrName = "User email or name",
  GroupName = "Group name",
}
