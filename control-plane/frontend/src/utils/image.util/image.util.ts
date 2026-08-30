export const imageUtil = {
  createBase64DataUrl,
};

function createBase64DataUrl(logoData: string | undefined | null): string | undefined {
  if (!logoData) return undefined;

  if (logoData.startsWith("data:")) {
    return logoData;
  }

  return `data:image/png;base64,${logoData}`;
}
