<template>
  <section class="tolerations-section">
    <div>Set toleration(s) to let the workload be scheduled on a node with a matching taint</div>
    <toleration-box
      v-for="item in displayedItems"
      :key="item.id"
      :toleration="item.origin"
      @remove-toleration="removeToleration(item.id)"
      @update-toleration="updateToleration(item.id, $event)"
      :read-only="item.isLocked || readOnly"
      :disabled-tooptip="item.isLocked ? disabledByPolicyText : undefined"
      :policy-rules="policyRules?.attributes"
      :allow-any-effect="allowAnyEffect"
      :read-only-info="disabledInfo"
    />
    <q-btn
      aid="add-toleration-btn"
      flat
      class="q-mt-sm btn-icon-sm"
      icon="ph-bold ph-plus"
      label="toleration"
      color="primary"
      @click="addToleration"
      :disable="cannotAddDueToPolicy || readOnly"
    >
      <q-tooltip v-if="addTolerationDisabledTooltip">{{ addTolerationDisabledTooltip }}</q-tooltip></q-btn
    >
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Components
import { TolerationBox } from "./toleration-box";
// Models
import type { Toleration } from "@/swagger-models/assets-service-client";
import { useItemizedPolicy } from "@/composables/use-itemized-policy.composable";
import { ItemizedMapperEntities } from "@/composables/helpers/use-itemized-policy.mapper";
import type { TolerationsDefaults, TolerationsRules } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "tolerations-section",
  components: {
    TolerationBox,
  },
  emits: ["update-tolerations"],
  props: {
    tolerations: {
      type: Array as PropType<Array<Toleration>>,
      default: () => [],
    },
    readOnly: {
      type: Boolean,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<TolerationsRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<TolerationsDefaults | null>,
      required: false,
    },
    allowAnyEffect: {
      type: Boolean,
      default: false,
    },
    disabledInfo: {
      type: String as PropType<string>,
      required: false,
    },
  },
  setup(props) {
    const { disabledByPolicyText, cannotAddDueToPolicy, addItem, removeItem, updateItem, displayedItems, fullList } =
      useItemizedPolicy(
        props.tolerations,
        ItemizedMapperEntities.Tolerations,
        props.policyRules?.instances,
        props.policyDefaults?.instances,
      );

    return {
      disabledByPolicyText,
      cannotAddDueToPolicy,
      addItem,
      removeItem,
      updateItem,
      displayedItems,
      fullList,
    };
  },
  computed: {
    addTolerationDisabledTooltip(): string | undefined {
      if (this.readOnly && this.disabledInfo) return this.disabledInfo;
      if (this.cannotAddDueToPolicy) return this.disabledByPolicyText;

      return undefined;
    },
  },
  methods: {
    getDefaultToleration(): Toleration {
      if (!this.policyDefaults?.attributes && !this.policyRules?.attributes) {
        return {
          key: "",
          operator: "Exists",
          value: null,
          effect: "NoSchedule",
        };
      }

      return {
        key: "",
        operator: null,
        value: null,
        effect: null,
        ...(this.policyDefaults?.attributes || {}),
      };
    },
    addToleration() {
      this.addItem(this.getDefaultToleration());
      this.$emit("update-tolerations", this.fullList);
    },
    updateToleration(id: string, toleration: Toleration) {
      this.updateItem(id, toleration);
      this.$emit("update-tolerations", this.fullList);
    },
    removeToleration(id: string) {
      this.removeItem(id);
      this.$emit("update-tolerations", this.fullList);
    },
  },
});
</script>
