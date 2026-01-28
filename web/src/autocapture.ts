// This feature is a work in progress. Still brainstorming the best approach.
export function initAutoCapture(
  trackFn: (name: string, props?: object) => void,
) {
  window.addEventListener(
    "click",
    (event) => {
      const target = event.target as HTMLElement;
      const element = target.closest(
        'button, a, input[type="submit"], [role="button"]',
      );

      if (!element) return;

      if (element.classList.contains("iris-ignore")) return;

      if (element instanceof HTMLInputElement && element.type === "password")
        return;

      const tagName = element.tagName.toLowerCase();
      let eventName = "$click";

      const props: Record<string, any> = {
        $tag: tagName,
        $id: element.id || undefined,
        $class: element.className || undefined,
        $text: (element as HTMLElement).innerText?.slice(0, 50) || "",
      };

      if (tagName === "a") {
        const href = (element as HTMLAnchorElement).getAttribute("href");
        if (href) props.$href = href;
      }

      trackFn(eventName, props);
    },
    { capture: true },
  );
}
