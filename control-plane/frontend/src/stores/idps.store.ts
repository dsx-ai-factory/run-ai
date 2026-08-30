import { defineStore } from "pinia";
import type { Idp, IdpCreationRequest, IdpResources, Mappers } from "@/swagger-models/identity-manager-client";
import { idpService } from "@/services/control-plane/idp.service/idp.service";

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
  },
  actions: {
    async loadIdps(): Promise<Array<Idp>> {
      this.idps = await idpService.getIdps();
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
