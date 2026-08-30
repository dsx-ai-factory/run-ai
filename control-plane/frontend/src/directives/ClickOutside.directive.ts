interface ExtendedHTMLElement extends HTMLElement {
  clickOutside: (event: MouseEvent) => void;
}

export const clickOutside = {
  mounted(el: ExtendedHTMLElement, { value }: { value: () => void }) {
    el.clickOutside = ({ clientX, clientY }: { clientX: number; clientY: number }) => {
      const { left, top, width, height } = el.getBoundingClientRect();
      if (!(clientX > left && clientX < left + width && clientY > top && clientY < top + height)) {
        value();
      }
    };
    document.addEventListener("click", el.clickOutside);
  },
  unmounted(el: ExtendedHTMLElement) {
    document.removeEventListener("click", el.clickOutside);
  },
};
