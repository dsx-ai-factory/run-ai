import type { IAccessInput } from "@/components/environment/connection-access/access-input-list";
import type { EToolLabel } from "./connection-access-modal.model";

export { convertNamesToItems };

function convertNamesToItems(names: string[], label: EToolLabel): IAccessInput[] {
  const accessInputs: IAccessInput[] = names.map((name: string) => ({
    text: name,
    label: label.toString(),
    removable: true,
  }));

  if (accessInputs.length === 1) {
    accessInputs[0].removable = false;
  }

  return accessInputs;
}
