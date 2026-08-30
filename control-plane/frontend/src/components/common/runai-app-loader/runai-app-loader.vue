<template>
  <section class="runai-app-loader">
    <vue3-lottie class="lottie-loader" :animation-data="modifiedAnimationData" :height="70" :width="70" />
  </section>
</template>

<script lang="ts">
import { Vue3Lottie } from "vue3-lottie";
import { defineComponent } from "vue";
import animationData from "./animation-data.json";
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";

const LOADING_TIMEOUT_MS = 90 * 1000;

export interface AnimationShape {
  ty: string;
  c?: {
    k: number[];
  };
  it?: AnimationShape[];
}

export interface AnimationLayer {
  shapes?: AnimationShape[];
}

export interface AnimationData {
  layers: AnimationLayer[];
}

export default defineComponent({
  components: {
    Vue3Lottie,
  },
  data() {
    return {
      timeoutId: null as ReturnType<typeof setTimeout> | null,
    };
  },
  mounted() {
    this.timeoutId = setTimeout(() => {
      this.$router.replace({ name: ERROR_ROUTE_NAMES.UNKNOWN_TIMEOUT });
    }, LOADING_TIMEOUT_MS);
  },
  beforeUnmount() {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
  },
  computed: {
    effectiveColor(): string {
      const cssVar = getComputedStyle(document.documentElement).getPropertyValue("--primary-color").trim();
      return cssVar || "#65a800";
    },
    modifiedAnimationData(): AnimationData {
      // Create a deep copy of the animation data
      const modifiedData = JSON.parse(JSON.stringify(animationData)) as AnimationData;

      // Use the effective color
      const color = this.effectiveColor.startsWith("#") ? this.effectiveColor : `#${this.effectiveColor}`;
      const rgb = this.hexToRgb(color);

      // Update all stroke and fill colors in the animation
      modifiedData.layers.forEach((layer: AnimationLayer) => {
        if (layer.shapes) {
          layer.shapes.forEach((shape: AnimationShape) => {
            if (shape.it) {
              shape.it.forEach((item: AnimationShape) => {
                if (item.ty === "fl" && item.c) {
                  item.c.k = [rgb.r / 255, rgb.g / 255, rgb.b / 255, 1];
                }
              });
            }
          });
        }
      });

      return modifiedData;
    },
  },
  methods: {
    hexToRgb(hex: string) {
      const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
      return result
        ? {
            r: parseInt(result[1], 16),
            g: parseInt(result[2], 16),
            b: parseInt(result[3], 16),
          }
        : { r: 0, g: 71, b: 255 }; // Default blue color
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-app-loader {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  bottom: 0;
  min-height: 100vh;
  .lottie-loader {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }
}
</style>
