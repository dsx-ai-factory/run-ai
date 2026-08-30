import type { TRolesTypes } from "./roles.model";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";

export interface ServiceAccountWithRoles extends ServiceAccount {
  roles?: TRolesTypes[];
}
