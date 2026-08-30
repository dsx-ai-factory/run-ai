import { policyUtil } from "./policy.util";
export { policyUtil };

// Export policy field iterator utilities for direct usage
export {
  clearWorkloadBlockedFields,
  isFieldBlocked,
  getBlockedFieldPaths,
  getBlockedFields,
  getValueByPath,
  deleteValueByPath,
  type IBlockableRule,
  type IBlockedFieldInfo,
  type TPolicyWithRules,
} from "./policy-field-iterator.util";
