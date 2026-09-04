/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import VueDOMPurifyHTML from "vue-dompurify-html";

const VueDOMPurifyHTMLOptions = {
  hooks: {
    // we want to enable open in a new tab when clicking on links
    afterSanitizeAttributes: (currentNode: HTMLElement) => {
      if ("target" in currentNode) {
        currentNode.setAttribute("target", "_blank");
      }
      // set non-HTML/MathML links to xlink:show=new
      if (
        !currentNode.hasAttribute("target") &&
        (currentNode.hasAttribute("xlink:href") || currentNode.hasAttribute("href"))
      ) {
        currentNode.setAttribute("xlink:show", "new");
      }
    },
  },
};

export { VueDOMPurifyHTML, VueDOMPurifyHTMLOptions };
