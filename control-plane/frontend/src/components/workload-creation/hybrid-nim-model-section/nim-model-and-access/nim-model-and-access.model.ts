import type { INimAccessModelModel } from "../nim-access-model/nim-access-model.model";

export interface INimModelAndAccessModel {
  modelName: string;
  profileNameOrHash: string;
  accessModel: INimAccessModelModel;
}
