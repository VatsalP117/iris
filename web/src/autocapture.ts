export function initAutoCapture(
    trackFn: (name: string, props?: object) => void,
): () => void {
    const handler = (event: Event) => {
        const target = event.target as HTMLElement;
        const element = target.closest(
            'button, a, input[type="submit"], [role="button"]',
        );

        if (!element) return;

        if (element.classList.contains("iris-ignore")) return;

        if (element instanceof HTMLInputElement && element.type === "password")
            return;

        const tagName = element.tagName.toLowerCase();
        const eventName = "$click";

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
    };

    window.addEventListener("click", handler, { capture: true });
    return () => window.removeEventListener("click", handler, { capture: true });
}
