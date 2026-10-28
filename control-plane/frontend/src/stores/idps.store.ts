/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import type { Idp, IdpCreationRequest, IdpResources, Mappers, SamlData } from "@/swagger-models/identity-manager-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { idpService } from "@/services/control-plane/idp.service/idp.service";
import { usePermissionStore } from "@/stores/permissions.store";

// TEMP: `samlData.validateSignature` is not in the generated identity-manager client yet. Remove this type (and use
// SamlData directly) once the backend API change that exposes it on GET /idps is merged and the client is regenerated.
type SamlDataWithValidateSignature = SamlData & { validateSignature?: boolean };

// TEMP (dev only): lets us see the SAML signature warnings before the backend change is merged.
// Enable from the browser console with: localStorage.setItem("mock-saml-validate-signature", "false")
// Remove together with SamlDataWithValidateSignature.
const MOCK_VALIDATE_SIGNATURE_KEY = "mock-saml-validate-signature";

function applyMockValidateSignature(idps: Array<Idp>): Array<Idp> {
  if (!import.meta.env.DEV || localStorage.getItem(MOCK_VALIDATE_SIGNATURE_KEY) !== "false") return idps;
  return idps.map((idp) =>
    idp.type === IdpTypeEnum.Saml
      ? { ...idp, samlData: { ...idp.samlData, validateSignature: false } as SamlDataWithValidateSignature }
      : idp,
  );
}

export const useIdpsStore = defineStore("Idps", {
  state: () => ({
    idps: new Array<Idp>(),
    idpMappers: {} as Mappers | undefined,
    idpResources: {} as IdpResources,
  }),
  getters: {
    currentIdp(): Idp | null {
      return this.idps.length > 0 ? this.idps[0] : null;
    },
    // Strict `=== false`: a backend that doesn't return the field yet must not trigger the warnings.
    isSamlSignatureValidationDisabled(): boolean {
      const samlData = this.currentIdp?.samlData as SamlDataWithValidateSignature | undefined;
      return this.currentIdp?.type === IdpTypeEnum.Saml && samlData?.validateSignature === false;
    },
    // The top banner is for admins who can fix it, so it is gated on the permission that allows editing the IdP.
    shouldShowSamlSignatureBanner(): boolean {
      return (
        this.isSamlSignatureValidationDisabled &&
        usePermissionStore().hasPermission(ResourceType.SecuritySettings, Action.Update)
      );
    },
  },
  actions: {
    async loadIdps(): Promise<Array<Idp>> {
      this.idps = applyMockValidateSignature(await idpService.getIdps());
      if (this.currentIdp) {
        this.idpMappers = this.currentIdp.mappers;
      } else {
        this.idpResources = {} as IdpResources;
        this.idpMappers = {} as Mappers;
      }
      return this.idps;
    },
    async loadIdpResources(): Promise<IdpResources> {
      if (this.currentIdp?.alias) {
        try {
          this.idpResources = await idpService.getIdpResource(this.currentIdp.alias);
          return this.idpResources;
        } catch (error: unknown) {
          return {};
        }
      }
      return {};
    },
    async removeIdp(idp: string): Promise<void> {
      await idpService.removeIdp(idp);
      await this.loadIdps();
    },
    async addIdp(idpCreationRequest: IdpCreationRequest): Promise<void> {
      if (!idpCreationRequest) return;
      await idpService.addIdp(idpCreationRequest);
      await this.loadIdps();
    },
    async updateIdp(idpAlias: string, idpCreationRequest: IdpCreationRequest): Promise<void> {
      if (!idpAlias || !idpCreationRequest) return;
      await idpService.updateIdp(idpAlias, idpCreationRequest);
      await this.loadIdps();
    },
  },
});
