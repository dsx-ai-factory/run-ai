import type { ITableColumn } from "@/models/table.model";

export interface IService {
  name: string;
}

export interface IServiceTableModalOptions {
  services: IService[];
  missingPrerequisites: string[];
  columns: Array<ITableColumn>;
  header: string;
  loading: boolean;
}
