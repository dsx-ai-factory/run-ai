/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import forge from 'node-forge';

export interface SamlIdpMetadataFixture {
    fileName: string;
    buffer: Buffer;
}

interface CertificateOptions {
    commonName: string;
    notBefore: Date;
    notAfter: Date;
}

function daysFromNow(days: number): Date {
    return new Date(Date.now() + days * 24 * 60 * 60 * 1000);
}

function generateSelfSignedCertificatePem(options: CertificateOptions): string {
    const keys = forge.pki.rsa.generateKeyPair(2048);
    const cert = forge.pki.createCertificate();
    cert.publicKey = keys.publicKey;
    cert.serialNumber = '01';
    cert.validity.notBefore = options.notBefore;
    cert.validity.notAfter = options.notAfter;
    const attrs = [{ name: 'commonName', value: options.commonName }];
    cert.setSubject(attrs);
    cert.setIssuer(attrs);
    cert.sign(keys.privateKey, forge.md.sha256.create());
    return forge.pki.certificateToPem(cert);
}

function certificatePemToBase64(pem: string): string {
    return pem
        .split('\n')
        .filter((line) => !line.includes('-----'))
        .join('');
}

function buildMetadataXml(certificateBase64: string): string {
    return `<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://e2e-saml-idp.example.com/saml/metadata">
  <md:IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:KeyDescriptor use="signing">
      <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:X509Data>
          <ds:X509Certificate>${certificateBase64}</ds:X509Certificate>
        </ds:X509Data>
      </ds:KeyInfo>
    </md:KeyDescriptor>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://e2e-saml-idp.example.com/saml/sso"/>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://e2e-saml-idp.example.com/saml/sso"/>
  </md:IDPSSODescriptor>
</md:EntityDescriptor>
`;
}

function toFixture(fileName: string, certificateBase64: string): SamlIdpMetadataFixture {
    return { fileName, buffer: Buffer.from(buildMetadataXml(certificateBase64), 'utf-8') };
}

// Certificates are generated at call time, relative to Date.now(), so the
// expiring-soon/expired windows stay correct no matter when the suite runs -
// a fixture checked into git with a hardcoded date would silently go stale.
export const samlCertificateFixtures = {
    // Valid for ~13 months: comfortably outside the 60-day "expiring soon" window.
    valid(): SamlIdpMetadataFixture {
        const pem = generateSelfSignedCertificatePem({
            commonName: 'e2e-saml-valid',
            notBefore: daysFromNow(-1),
            notAfter: daysFromNow(400)
        });
        return toFixture('idp-metadata-valid.xml', certificatePemToBase64(pem));
    },
    // Expires in 10 days: inside the 60-day "expiring soon" threshold, but still valid.
    expiringSoon(): SamlIdpMetadataFixture {
        const pem = generateSelfSignedCertificatePem({
            commonName: 'e2e-saml-expiring-soon',
            notBefore: daysFromNow(-1),
            notAfter: daysFromNow(10)
        });
        return toFixture('idp-metadata-expiring-soon.xml', certificatePemToBase64(pem));
    },
    // Well-formed certificate whose validity window is already in the past.
    expired(): SamlIdpMetadataFixture {
        const pem = generateSelfSignedCertificatePem({
            commonName: 'e2e-saml-expired',
            notBefore: daysFromNow(-400),
            notAfter: daysFromNow(-30)
        });
        return toFixture('idp-metadata-expired.xml', certificatePemToBase64(pem));
    },
    // Truncated certificate data: fails X.509 parsing entirely, distinct from "expired".
    malformed(): SamlIdpMetadataFixture {
        const pem = generateSelfSignedCertificatePem({
            commonName: 'e2e-saml-malformed',
            notBefore: daysFromNow(-1),
            notAfter: daysFromNow(400)
        });
        const truncatedCertificateBase64 = certificatePemToBase64(pem).slice(0, 40);
        return toFixture('idp-metadata-malformed.xml', truncatedCertificateBase64);
    }
};
