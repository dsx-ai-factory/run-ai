import forge from "node-forge";

export const certificateUtil = {
  getExpirationDate,
};

async function getExpirationDate(base64Cert: string): Promise<Date | null> {
  try {
    const certData = forge.util.decode64(base64Cert);
    const cert = forge.pki.certificateFromAsn1(forge.asn1.fromDer(certData));
    return cert.validity.notAfter;
  } catch (e: unknown) {
    console.warn("Failed to get expiration date from certificate: ", (e as Error)?.message);
    return null;
  }
}
