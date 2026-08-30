import type { IMenuItem, ISubjectRoutes } from "@/components/infra/runai-layout/routes-model/routes-model";

export const runaiHeaderUtil = {
  findLabelPath,
};

function findLabelPath(allSubjects: IMenuItem[], targetLabel: string) {
  for (const subject of allSubjects) {
    const result = search(subject.items, targetLabel);
    if (result) {
      // Extract just the subject label without any subtitle
      const subjectLabel = subject.label.split("\n")[0];
      return [subjectLabel, ...result];
    }
  }

  return null; // If not found in any subjects
}

function search(items: ISubjectRoutes[], targetLabel: string, path: string[] = []): string[] | null {
  for (const item of items) {
    // For the path, use the full pageTitle (with subtitle) to preserve it
    const currentPath = [...path, targetLabel.startsWith(item.label + " (") ? targetLabel : item.label];

    // Check if the label matches (exact match or targetLabel starts with item.label)
    // This handles cases like "Service accounts (formerly applications)" matching "Service accounts"
    const isMatch = item.label === targetLabel || targetLabel.startsWith(item.label + " (");

    if (isMatch) {
      return currentPath;
    }

    // Search in subItems or items recursively
    const subItems = item.subItems || item.subItems;
    if (subItems) {
      const result = search(subItems, targetLabel, currentPath);
      if (result) {
        return result;
      }
    }
  }
  return null; // If not found in current items
}
