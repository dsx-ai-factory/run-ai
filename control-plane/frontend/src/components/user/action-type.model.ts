export enum ActionTypeEnum {
  CREATE = "create",
  RESET = "reset",
  REGENERATE = "regenerate",
  DONE = "done",
}

export type TActionTypes = ActionTypeEnum.CREATE | ActionTypeEnum.RESET;
