import type { ArrayRules } from "@/swagger-models/policy-service-client";

export enum ENodePoolsListOrigin {
  Project = "project",
  Template = "template",
}

export interface IComputeSectionNodePoolsPolicy {
  rules?: ArrayRules;
  defaults?: Array<string>;
}
