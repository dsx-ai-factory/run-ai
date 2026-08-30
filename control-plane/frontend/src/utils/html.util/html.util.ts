export const HTMLUtil = {
  HTMLToPlainText,
  appendSpacesToHtml,
  downloadFileFromString,
};

function HTMLToPlainText(html: string): string {
  const noTags = html.replace(/<div\s*\/?>/gi, " ").replace(/<[^>]*>/g, "");
  return noTags.replace(/&nbsp;/g, " ");
}

function appendSpacesToHtml(html: string, spaceCount: number): string {
  return html.concat("&nbsp;".repeat(spaceCount));
}

function downloadFileFromString(fileContent: string, fileName: string, type = "text/plain"): void {
  const blob: Blob = new Blob([fileContent], { type: type });
  const link: HTMLAnchorElement = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = fileName as string;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}
