import { ROOT_URL, getUrlByEnvType } from "@/constants/docs";
import type { ComponentOptions } from "vue";

// docsLinkMixin is a mixin that modifies the href of all <a> elements in the component's DOM to point to the correct version of the documentation.
// Since the app version is loaded after the app is mounted, we need to use a mixin to modify the href of the <a> elements.
// We have multiple a elements that are hard-coded in typescript files, so they are loaded before the version is loaded.
// Example:
// <a href="https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/system-requirements">System Requirements</a>
// should be modified to:
// <a href="https://run-ai-docs.nvidia.com/self-hosted/2.21/getting-started/installation/system-requirements">System Requirements</a>
// Example 2 (saas):
// <a href="https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/system-requirements">System Requirements</a>
// should be modified to:
// <a href="https://run-ai-docs.nvidia.com/saas/getting-started/installation/system-requirements">System Requirements</a>
export const docsLinkMixin: ComponentOptions = {
  mounted() {
    this.updateDocsLinks();
    this.setupMutationObserver();
  },
  beforeUnmount() {
    if (this.mutationObserver) {
      this.mutationObserver.disconnect();
    }
  },
  methods: {
    updateDocsLinks() {
      const element = this.$el;
      if (!element || !element.querySelectorAll) {
        return;
      }

      const anchors = element.querySelectorAll("a[href]");
      anchors.forEach((el: HTMLElement) => {
        this.processAnchorElement(el);
      });
    },

    processAnchorElement(el: HTMLElement) {
      const originalLink = el.getAttribute("href");
      if (!originalLink) return;
      if (!originalLink.startsWith(ROOT_URL)) return;

      el.setAttribute("href", getUrlByEnvType(originalLink));
    },

    setupMutationObserver() {
      const element = this.$el;
      if (!element) return;

      this.mutationObserver = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
          if (mutation.type !== "childList") return;
          mutation.addedNodes.forEach((node) => {
            if (node.nodeType !== Node.ELEMENT_NODE) return;

            const element = node as HTMLElement;
            // Check if the added node is an anchor element
            if (element.tagName === "A" && element.hasAttribute("href")) {
              this.processAnchorElement(element);
            }
            // Check for anchor elements within the added node
            const anchors = element.querySelectorAll?.("a[href]");
            if (anchors) {
              anchors.forEach((anchor: Element) => {
                this.processAnchorElement(anchor as HTMLElement);
              });
            }
          });
        });
      });

      this.mutationObserver.observe(element, {
        childList: true,
        subtree: true,
      });
    },
  },
};
